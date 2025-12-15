package github

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/google/go-github/v60/github"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "workctl/internal/handler/github"

type Handler struct {
	logger *slog.Logger
	secret []byte
	tracer trace.Tracer
}

type Option func(*Handler)

func WithLogger(logger *slog.Logger) Option {
	return func(h *Handler) {
		h.logger = logger
	}
}

func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(h *Handler) {
		h.tracer = tp.Tracer(instrumentationName)
	}
}

func NewHandler(opts ...Option) *Handler {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	h := &Handler{
		logger: slog.Default(),
		secret: []byte(secret),
		tracer: otel.Tracer(instrumentationName),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Extract context for distributed tracing propagation
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	// 2. Start Span
	ctx, span := h.tracer.Start(ctx, "github.process_webhook",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	eventType := github.WebHookType(r)
	deliveryID := github.DeliveryID(r)

	span.SetAttributes(
		attribute.String("github.event_type", eventType),
		attribute.String("github.delivery_id", deliveryID),
	)

	payload, err := github.ValidatePayload(r, h.secret)
	if err != nil {
		h.logger.Error("failed to validate webhook signature", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to validate webhook signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		h.logger.Error("failed to parse webhook payload", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse webhook payload")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Enrich span with metadata
	if evt, ok := event.(interface{ GetRepo() *github.Repository }); ok {
		if repo := evt.GetRepo(); repo != nil {
			span.SetAttributes(attribute.String("github.repository", repo.GetFullName()))
		}
	}
	if evt, ok := event.(interface{ GetSender() *github.User }); ok {
		if sender := evt.GetSender(); sender != nil {
			span.SetAttributes(attribute.String("github.sender", sender.GetLogin()))
		}
	}

	// Event-specific attributes
	switch e := event.(type) {
	case *github.PushEvent:
		span.SetAttributes(attribute.String("github.ref", e.GetRef()))
	case *github.PullRequestEvent:
		span.SetAttributes(attribute.Int("github.pr.number", e.GetNumber()))
	case *github.WorkflowJobEvent:
		span.SetAttributes(
			attribute.String("github.workflow_job.name", e.GetWorkflowJob().GetName()),
			attribute.Int64("github.workflow_job.id", e.GetWorkflowJob().GetID()),
		)
	}

	// Just for log correlation/debug
	h.logger.InfoContext(ctx, "webhook received", "event_type", eventType, "delivery_id", deliveryID)

	// Assuming processing happens here or is dispatched. for now just return OK.
	// If actual processing happened, we'd probably have child spans.

	w.WriteHeader(http.StatusOK)
}
