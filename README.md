# **BACOPA**

**Open Creative Multi-Game Playground Where Humans, Creators, Players, And Bots Meet**

## **What is BACOPA?**

*Bacopa monnieri* is an aquatic plant known for two things: its rapid, dense branching structure, and its historical use as a nootropic (cognitive enhancer).  
**BACOPA** the platform is exactly that:

1. **Rapid Branching:** A playground where game designers and enthusiasts can endlessly "fork" existing games, tweak rulesets, and invent entirely new variants.  
2. **Cognitive Enhancement:** Powered by the [**RLGB (Reinforcement Learning Game Bot)**](https://github.com/horkah/RLGB) engine, BACOPA automatically trains state-of-the-art AI agents to play these user-generated games using Proximal Policy Optimization (PPO), self-play, and evolutionary curriculum learning.

BACOPA bridges the gap between traditional gaming, tabletop design, and modern machine learning. It is a unified ecosystem where players can face off against human or bot opponents, and designers can instantly test their chaotic new mechanics against intelligent agents.

---

## **Current Status: v0.1 — Playable MVP**

The first working version of BACOPA is live with the following features:

### What's Working Now
- **Two Games:** Tic Tac Toe and Connect Four, fully playable
- **User Management:** Registration, login, JWT authentication, Elo ratings
- **Human vs Human:** Real-time multiplayer via WebSocket with in-game chat
- **Human vs AI:** Minimax AI with alpha-beta pruning at three difficulty levels (Easy, Medium, Hard)
- **Game Lobby:** Create games, browse open games, join matches
- **Match History:** Track past games with results and Elo changes
- **Dark-themed UI:** Modern gaming aesthetic with animations

### Architecture Decisions for v0.1
| **Component** | **Technology** | **Status** |
|---|---|---|
| Frontend | Next.js + TypeScript + Tailwind CSS | Implemented |
| Backend | Go (gorilla/mux + gorilla/websocket) | Implemented |
| Database | SQLite (WAL mode) | Implemented |
| AI Opponents | Minimax with alpha-beta pruning | Implemented |
| Real-time | WebSockets (native gorilla/websocket) | Implemented |
| Containerization | Docker + Docker Compose | Configured |

**Key v0.1 decisions:**
- AI uses classical minimax instead of RLGB — sufficient for simple games, no GPU needed
- SQLite instead of PostgreSQL — zero-config, suitable for single-server deployment
- Monolithic Go backend instead of microservices — simpler for two games, will split later
- No rule editor — games are hardcoded; the compiler layer comes in Phase 3

---

## **Roadmap**

### Phase 1 — Playable MVP ✅ (Current)
- Two hardcoded games (Tic Tac Toe, Connect Four)
- User auth, Elo ratings, match history
- PvP with chat, AI opponents (minimax)

### Phase 2 — Community & Polish
- Game spectating / live observation
- Leaderboards and detailed player profiles
- Game replay system
- Lobby chat and matchmaking queue
- More games (Checkers, Reversi, etc.)

### Phase 3 — Rule Editor & Compiler
- Visual drag-and-drop game editor
- Universal JSON ruleset schema
- Compiler layer: JSON → RLGB Python Game class
- Fuzz testing for user-created rulesets

### Phase 4 — ML Pipeline & RLGB Integration
- FastAPI inference service wrapping RLGB
- Auto-training pipeline (Celery + GPU pods)
- Self-play PPO training for user-created games
- Model registry and versioning

### Phase 5 — Scale & Production
- Kubernetes orchestration
- Redis for session state caching
- PostgreSQL migration for multi-server deployment
- CDN and production deployment
- Rate limiting, abuse prevention

---

## **The Technology Stack (Full Vision)**

| **Component** | **Technology** | **Why** |
|---|---|---|
| **Frontend** | Next.js + TypeScript + Tailwind | Fast loads, type safety, modern UI |
| **Game Renderer** | PixiJS (2D) / Three.js (3D) | Dynamic board rendering from JSON state (Phase 3+) |
| **Platform Backend** | Go | High concurrency for WebSocket routing, matchmaking, game sessions |
| **Live Inference API** | Python (FastAPI) | Wraps RLGB engine for real-time AI moves (Phase 4+) |
| **Real-time** | WebSockets | Multiplayer moves, chat, live observation |
| **Database** | SQLite → PostgreSQL | SQLite now, PostgreSQL when scaling (Phase 5) |
| **MLOps** | Docker, Kubernetes, Celery | GPU pod orchestration for auto-training (Phase 4+) |

## **Core Architecture & Design Requirements**

BACOPA's architecture is split into three highly distinct domains.

### **1. The Live Arena (State Management & Inference)**

RLGB's core design dictates that games are stateless and transition functionally (next\_state(state, action)).

* The **Platform Backend** (Go) holds the "ground truth" of any active multiplayer game.  
* When it is a Bot's turn, the Platform Backend either runs local minimax (v0.1) or sends the current state via HTTP to the **Inference API** (FastAPI, Phase 4+).
* FastAPI dynamically loads the corresponding PyTorch model for that specific game variant, runs agent.act(), and returns the move.

### **2. The Universal Compiler (The Grand Challenge)**

Users create games in a visual drag-and-drop web editor. RLGB expects a strict Python Game subclass with explicit mathematical action spaces and perspective-relative observe(state) tensors.

* BACOPA must include a **Compiler Layer**.  
* This layer accepts a universal JSON/YAML schema from the frontend editor and automatically translates it into a valid RLGB Python class, bridging visual mechanics with neural network tensor inputs.
* **Status:** Not yet started (Phase 3)

### **3. Asynchronous MLOps (The Training Pipeline)**

When a user invents or forks a game and clicks "Publish":

* A message is sent to a job queue (Celery/RabbitMQ).  
* Kubernetes spins up a GPU training pod.  
* The pod executes RLGB's run\_optimize.py to train a baseline bot via self-play.  
* Once training yields a competent agent, the model is saved and the game goes "Live" with an AI opponent.
* **Status:** Not yet started (Phase 4)

---

## **Project Structure**

```
BACOPA/
├── backend-platform/       # Go backend server
│   ├── cmd/server/         # Entry point
│   ├── internal/
│   │   ├── ai/             # Minimax AI with alpha-beta pruning
│   │   ├── auth/           # JWT authentication
│   │   ├── db/             # SQLite database layer
│   │   ├── game/           # Game engines (TicTacToe, ConnectFour)
│   │   ├── handler/        # HTTP & WebSocket handlers
│   │   └── models/         # Data models
│   ├── Dockerfile
│   └── go.mod
├── frontend/               # Next.js frontend
│   ├── src/
│   │   ├── app/            # Pages (login, register, game, dashboard)
│   │   ├── components/     # UI components (boards, chat, modals)
│   │   ├── context/        # Auth context
│   │   ├── hooks/          # WebSocket hook
│   │   └── lib/            # API client
│   ├── Dockerfile
│   └── package.json
├── RLGB/                   # RLGB engine (git submodule)
├── docker-compose.yml
└── README.md
```

---

## **Instructions for the Development Team**

### **A. Topic Areas / Responsibilities**

1. **The Platform (Go & DB):** User authentication, Elo rating algorithms, WebSocket routing, game lobby state management, and database schemas.
2. **The Compiler (Python):** Designs the universal JSON Ruleset Schema and builds the parser that turns frontend JSON into robust RLGB Game classes.
3. **The MLOps & Inference (Python):** Wraps RLGB into the FastAPI service for live play and orchestrates automated GPU training queues.
4. **The Interface & Editor (React):** Builds the player dashboards, community features, and the visual drag-and-drop game editor.

### **B. Development Guidelines**

* **API-First Development:** OpenAPI/Swagger specs defining game states and ruleset structures must be agreed upon before building.
* **Strict Typing & Validation:** TypeScript on frontend, strong types in Go, Pydantic in Python.
* **Fuzz Testing:** The Compiler must handle random/garbage JSON rulesets gracefully.
* **Trunk-Based Development:** Small, frequent updates to main behind feature flags.

---

## **Getting Started (Local Development)**

### **Prerequisites**

* Go 1.21+ (with CGO support for SQLite)
* Node.js 20+
* Docker & Docker Compose (optional)

### **Quick Start**

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/horkah/bacopa.git
cd bacopa

# Start backend (terminal 1)
cd backend-platform
go mod tidy
go run ./cmd/server/

# Start frontend (terminal 2)
cd frontend
npm install
npm run dev
```

Open http://localhost:3000 in your browser.

### **Docker Compose**

```bash
docker-compose up --build
```

This starts the backend on port 8080 and the frontend on port 3000.

### **Development Workflow**

1. Register an account at http://localhost:3000/register
2. Choose a game (Tic Tac Toe or Connect Four)
3. Play vs AI (Easy/Medium/Hard) or create a PvP game
4. For PvP: share the game link with another player, or have them join from the lobby

---

## **API Reference**

### REST Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/auth/register` | No | Create account |
| POST | `/api/auth/login` | No | Login |
| GET | `/api/auth/me` | Yes | Get current user |
| GET | `/api/games/types` | No | List available games |
| POST | `/api/games` | Yes | Create game session |
| GET | `/api/games/lobby` | Yes | List open PvP games |
| POST | `/api/games/:id/join` | Yes | Join a PvP game |
| GET | `/api/games/history` | Yes | Match history |

### WebSocket

Connect to `ws://localhost:8080/ws?token=JWT&gameId=ID`

**Client messages:** `move` (position), `chat` (message)  
**Server messages:** `game_state`, `chat`, `player_joined`, `game_over`, `error`

---

## **License: AGPLv3**

BACOPA is licensed under the **GNU Affero General Public License v3.0**. The AGPLv3 closes the "Application Service Provider" loophole — if anyone takes the BACOPA codebase, modifies it, and hosts it as a web service, they must release their modifications back to the open-source community.
