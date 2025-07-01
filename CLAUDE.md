# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Bitfinex lending bot written in Go that automates margin lending strategies. The bot:
- Monitors lending rates and automatically places optimized lending offers
- Implements sophisticated multi-tier lending strategies with dynamic rate adjustments
- Provides Telegram bot integration for real-time monitoring and configuration
- Supports high-hold strategies for large amounts at premium rates
- Includes rate-based period selection (2-day, 30-day, 120-day loans)

## Architecture

**Single-file application**: All code is in `main.go` (780 lines) - no separate packages or modules.

**Core components**:
- `envStruct`: Configuration management using Viper (YAML config + env vars)
- `MarginBotConf`: Runtime lending strategy parameters
- `MarginBotLoanOffer`: Individual loan offer structure
- Bitfinex API v2 REST client integration for funding operations
- Telegram bot for remote monitoring and configuration

**Key algorithms**:
- `marginBotGetLoanOffers()`: Core strategy engine that calculates optimal loan offers based on market depth, available funds, and configured parameters
- Rate ladder algorithm using funding book positions (`GAP_BOTTOM` to `GAP_TOP`)
- High-hold strategy for premium rate lending on large amounts
- Dynamic period selection based on rate thresholds

## Development Commands

**Build and run**:
```bash
go run main.go                    # Run with default config.yaml
go run main.go -c custom.yaml     # Run with custom config file
go build -o BitfinexLendingBot    # Build executable
```

**Code formatting and validation**:
```bash
gofmt -w main.go                  # Format code
go vet                            # Static analysis
go mod tidy                       # Clean up dependencies
go mod vendor                     # Update vendor directory
```

**Testing**:
No test files exist in this codebase. The application appears to be tested manually against live Bitfinex API.

## Configuration

**Primary config**: `config.yaml` contains all bot parameters including API keys, lending strategies, and Telegram settings.

**Critical settings**:
- `BITFINEX_API_KEY` / `BITFINEX_SECRET_KEY`: Trading credentials
- `MIN_DAILY_LEND_RATE`: Minimum acceptable lending rate (safety threshold)
- `ORDER_LIMIT`: Maximum orders per execution cycle (risk management)
- `SPREAD_LEND`: Number of orders to split funds across
- `GAP_BOTTOM` / `GAP_TOP`: Market depth range for rate calculation
- `HIGH_HOLD_*`: Premium lending strategy for large amounts

## Key Functions Reference

**Main execution flow**:
- `botRun()` at main.go:182 - Main bot execution cycle
- `marginBotGetLoanOffers()` at main.go:308 - Strategy calculation engine
- `placeLoanOffers()` at main.go:285 - Order placement logic

**Telegram integration**:
- `handleTelegramMessages()` at main.go:434 - Command processing
- Dynamic configuration updates via chat commands
- Rate monitoring with threshold alerts

**API operations**:
- `cancelAllOffers()` at main.go:245 - Funding offer management using v2 API
- `getAvailableFunds()` at main.go:274 - Wallet balance checking using v2 API
- `getLendRate()` at main.go:733 - Current funding rate retrieval using v2 API
- Integration with bitfinex-api-go v2 REST library

**V2 API Key Changes**:
- Funding symbols use "f" prefix (e.g., "fUSD" instead of "USD")
- Funding offers accessed via `client.Funding.Offers(symbol)`
- Wallet balances accessed via `client.Wallet.Wallet()`
- Funding book accessed via `client.Book.All(symbol, precision, limit)`
- Rate values are daily rates (not annual) in v2 API

## Security Notes

- API credentials are stored in config.yaml (ensure this file is not committed)
- Telegram bot token is in config.yaml
- Single chat ID authentication for Telegram access
- No input validation on CLI arguments - config file path is trusted