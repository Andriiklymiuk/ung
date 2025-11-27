# UNG

**Stop losing money on unbilled hours.**

A CLI tool that tracks your time, manages clients, and generates invoices. Works offline. Your data stays on your machine.

## Why UNG?

- **No subscriptions** — free forever, no SaaS fees eating your profits
- **No cloud** — your client data never leaves your computer
- **No friction** — start a timer in 2 seconds, invoice in 10

## Install

```bash
brew install andriiklymiuk/tools/ung
```

## 60-Second Setup

```bash
ung config init --global
ung company add --name "Your Name" --email "you@email.com"
ung client add --name "Acme Corp" --email "billing@acme.com"
ung contract add --client 1 --name "Development" --type hourly --rate 100
```

## Track Time

```bash
ung track start --contract 1      # Start working
ung track stop                    # Done for now
ung track ls                      # See your hours
```

## Get Paid

```bash
ung invoice --client "Acme Corp"  # Create from tracked time
ung invoice pdf 1                 # Generate PDF
ung invoice email 1               # Send it
```

## Commands

| | |
|---|---|
| `ung create` | Interactive wizard |
| `ung track now` | Current timer |
| `ung dashboard` | Revenue overview |
| `ung invoice ls` | All invoices |
| `ung doctor` | Health check |

## VS Code Extension

Search **"UNG"** in VS Code Extensions for a visual interface.

## Links

[Docs](https://andriiklymiuk.github.io/ung) · [Issues](https://github.com/Andriiklymiuk/ung/issues)

---

MIT License · Built for freelancers who bill by the hour
