# Agent Work Batcher

The goal of this project is to handle automatic assignments to an agentic developer (e.g. Google Jules), so that we can
execute large chunks of work in series without necessarily checking in on it (like, overnight).

Right now, the developer must be checked on every 15 - 30 minutes. While fun, this is very distracting. This project aims
to batch multiple hours of work, and at the end of it, simply release it.

## Development

This project is written in Go and uses Cobra for the CLI.
