# **BACOPA**

**Open Creative Multi-Game Playground Where Humans, Creators, Players, And Bots Meet**

## **What is BACOPA?**

*Bacopa monnieri* is an aquatic plant known for two things: its rapid, dense branching structure, and its historical use as a nootropic (cognitive enhancer).  
**BACOPA** the platform is exactly that:

1. **Rapid Branching:** A playground where game designers and enthusiasts can endlessly "fork" existing games, tweak rulesets, and invent entirely new variants.  
2. **Cognitive Enhancement:** Powered by the [**RLGB (Reinforcement Learning Game Bot)**](https://github.com/horkah/RLGB) engine, BACOPA automatically trains state-of-the-art AI agents to play these user-generated games using Proximal Policy Optimization (PPO), self-play, and evolutionary curriculum learning.

BACOPA bridges the gap between traditional gaming, tabletop design, and modern machine learning. It is a unified ecosystem where players can face off against human or bot opponents, and designers can instantly test their chaotic new mechanics against intelligent agents.

---

## **Current Status: v0.2 — Unified Architecture**

### What's Working Now
- **Two Games:** Tic Tac Toe and Connect Four, fully playable
- **User Management:** Registration, login, JWT authentication, Elo ratings
- **Human vs Human:** Real-time multiplayer via WebSocket with in-game chat
- **Human vs AI:** Minimax AI with alpha-beta pruning at three difficulty levels (Easy, Medium, Hard)
- **Game Lobby:** Create games, browse open games, join matches
- **Match History:** Track past games with results and Elo changes
- **Dark-themed UI:** Modern gaming aesthetic with animations

### Architecture: Zero Redundancy Design

The v0.2 architecture follows a strict **single source of truth** principle:

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Frontend   │◄───►│   Go Backend     │◄───►│  RLGB Service   │
│  (Next.js)   │ WS  │  (Platform Layer)│ HTTP│  (Game Engine)  │
│              │     │                  │     │                 │
│ • UI/UX      │     │ • Auth (JWT)     │     │ • Game rules    │
│ • Game boards│     │ • WebSocket hub  │     │ • Move validation│
│ • Chat       │     │ • Elo ratings    │     │ • Win detection │
│ • Routing    │     │ • Game lobby     │     │ • AI (minimax)  │
│              │     │ • DB (SQLite)    │     │ • Serialization │
└─────────────┘     └──────────────────┘     └─────────────────┘
```

**Key design decisions:**

- **RLGB is the single source of truth** for all game logic. The Go backend contains zero game rules — it treats game state as an opaque JSON blob passed through to/from RLGB.
- **Adding a new game = one Python file in RLGB.** Zero Go or frontend changes needed (the frontend renders dynamically based on game type).
- **Minimax AI lives in RLGB** as a generic agent that works with any Game subclass via the Game interface. No game-specific AI code needed.
- **Clean service boundary:** Go owns the platform (auth, real-time, persistence), RLGB owns the game intelligence (rules, AI, training).

| **Component** | **Technology** | **Responsibility** |
|---|---|---|
| Frontend | Next.js + TypeScript + Tailwind | UI, game rendering, chat |
| Platform Backend | Go (gorilla/mux + gorilla/websocket) | Auth, WebSocket routing, Elo, lobby, DB |
| Game Engine | Python (RLGB + FastAPI) | Game rules, move validation, AI, serialization |
| Database | SQLite (WAL mode) | Users, game sessions (state stored opaquely) |
| Containerization | Docker + Docker Compose | Three-service deployment |

---

## **How It Works**

### Game Flow

1. Player creates a game → Go backend calls RLGB `/games/{type}/new` → gets initial state
2. Player connects via WebSocket → Go loads game room with opaque RLGB state
3. Player makes a move → Go calls RLGB `/games/{type}/move` with state + action → gets new state + display
4. If AI mode → Go calls RLGB `/games/{type}/ai-move` → minimax computes best move → returns new state
5. Go broadcasts display state to all connected WebSocket clients
6. On game over → Go updates Elo ratings and persists final state

### RLGB Game Interface

Every game in RLGB implements the same abstract interface:

```python
class Game(ABC):
    num_actions: int              # Size of action space
    observation_shape: tuple      # Tensor shape for neural networks
    num_players: int              # Number of players
    
    def initial_state() → State          # Starting position
    def current_player(state) → int      # Whose turn
    def legal_actions(state) → list[int] # Valid moves
    def next_state(state, action) → State # Apply move (immutable)
    def is_terminal(state) → bool        # Game over?
    def returns(state) → list[float]     # Per-player scores (+1/-1/0)
    def observe(state) → Tensor          # Neural network input
```

This means:
- **Minimax works generically** — it only uses `legal_actions()`, `next_state()`, `is_terminal()`, and `returns()`. Any game that implements the interface gets AI for free.
- **Neural agents** (PPO) can be trained via self-play on any game — they use `observe()` for tensor inputs and `legal_actions()` for action masking.
- **The same game class serves both online play and ML training** — no code duplication.

---

## **Roadmap**

### Phase 1 — Playable MVP ✅
- Two hardcoded games (Tic Tac Toe, Connect Four)
- User auth, Elo ratings, match history
- PvP with chat, AI opponents (minimax)

### Phase 1.5 — Unified Architecture ✅ (Current)
- RLGB as single source of truth for game logic
- Go backend as pure platform layer (zero game rules)
- Generic minimax agent in RLGB (works with any game)
- FastAPI game service with serialization layer
- Clean service boundary between platform and game engine

### Phase 2 — Community & More Games
- Game spectating / live observation
- Leaderboards and detailed player profiles
- Game replay system
- More games: Kuhn Poker, Tron (already in RLGB — just need frontend components)
- Neural AI agents (load trained RLGB models for inference)

### Phase 3 — Rule Editor & Compiler
- Visual drag-and-drop game editor
- Universal JSON ruleset schema
- Compiler layer: JSON → RLGB Python Game class
- Fuzz testing for user-created rulesets

### Phase 4 — ML Pipeline & Auto-Training
- Auto-training pipeline (Celery + GPU pods)
- Self-play PPO training for user-created games
- Model registry and versioning
- Curriculum learning with league training

### Phase 5 — Scale & Production
- Kubernetes orchestration
- Redis for session state caching
- PostgreSQL migration for multi-server deployment
- CDN and production deployment

---

## **Project Structure**

```
BACOPA/
├── RLGB/                       # Game engine (git submodule) — SINGLE SOURCE OF TRUTH
│   ├── rlgb/
│   │   ├── base.py             # Game + Agent abstract interfaces
│   │   ├── server.py           # FastAPI game service (port 9090)
│   │   ├── serialization.py    # State ↔ JSON ↔ frontend display conversion
│   │   ├── games/
│   │   │   ├── tictactoe.py    # Tic Tac Toe implementation
│   │   │   ├── connect_four.py # Connect Four implementation
│   │   │   ├── kuhn_poker.py   # Kuhn Poker (future)
│   │   │   └── tron.py         # 4-player Tron (future)
│   │   ├── agents/
│   │   │   ├── minimax.py      # Generic minimax with alpha-beta pruning
│   │   │   ├── neural.py       # PPO neural agent
│   │   │   └── random_agent.py # Random baseline
│   │   ├── trainer.py          # Self-play + league training loop
│   │   └── optimize.py         # Curriculum + evolutionary optimization
│   ├── models/                 # Trained neural network weights
│   └── Dockerfile
├── backend-platform/           # Platform layer (Go) — NO game logic
│   ├── cmd/server/main.go      # Entry point
│   ├── internal/
│   │   ├── rlgb/client.go      # HTTP client for RLGB service
│   │   ├── auth/auth.go        # JWT authentication
│   │   ├── db/db.go            # SQLite database layer
│   │   ├── handler/
│   │   │   ├── auth.go         # Register/login handlers
│   │   │   ├── game.go         # Game CRUD, lobby (delegates to RLGB)
│   │   │   └── ws.go           # WebSocket hub (delegates moves to RLGB)
│   │   └── models/models.go    # Data models
│   └── Dockerfile
├── frontend/                   # Web client (Next.js)
│   ├── src/
│   │   ├── app/                # Pages (login, register, game, dashboard)
│   │   ├── components/         # UI (TicTacToe, ConnectFour, Chat, modals)
│   │   ├── context/            # Auth context
│   │   ├── hooks/              # WebSocket hook
│   │   └── lib/                # API client
│   └── Dockerfile
├── docker-compose.yml          # Three-service orchestration
└── README.md
```

---

## **Getting Started (Local Development)**

### **Prerequisites**

* Go 1.21+ (with CGO for SQLite)
* Python 3.11+ with pip
* Node.js 20+
* Docker & Docker Compose (optional)

### **Quick Start (3 terminals)**

```bash
# Terminal 1: RLGB Game Service
cd RLGB
pip install -r requirements.txt
uvicorn rlgb.server:app --host 0.0.0.0 --port 9090

# Terminal 2: Go Platform Backend
cd backend-platform
go mod tidy
go run ./cmd/server/

# Terminal 3: Frontend
cd frontend
npm install
npm run dev
```

Open http://localhost:3000 in your browser.

### **Docker Compose (single command)**

```bash
docker-compose up --build
```

This starts RLGB (9090), backend (8080), and frontend (3000).

### **Playing**

1. Register at http://localhost:3000/register
2. Choose Tic Tac Toe or Connect Four from the dashboard
3. Play vs AI (Easy/Medium/Hard) or create a PvP game
4. For PvP: share the game link or have the opponent join from the lobby
5. Chat with your opponent in real-time during PvP games

---

## **API Reference**

### RLGB Game Service (port 9090)

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/games` | List available game types |
| POST | `/games/{type}/new` | Create initial game state |
| POST | `/games/{type}/move` | Validate and apply a move |
| POST | `/games/{type}/ai-move` | Get AI move and apply it |

### Platform Backend (port 8080)

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/auth/register` | No | Create account |
| POST | `/api/auth/login` | No | Login |
| GET | `/api/auth/me` | Yes | Get current user |
| GET | `/api/games/types` | No | List games (proxies to RLGB) |
| POST | `/api/games` | Yes | Create game session |
| GET | `/api/games/lobby` | Yes | List open PvP games |
| POST | `/api/games/:id/join` | Yes | Join a PvP game |
| GET | `/api/games/history` | Yes | Match history |

### WebSocket (port 8080)

Connect: `ws://localhost:8080/ws?token=JWT&gameId=ID`

**Client → Server:** `move` (position), `chat` (message)  
**Server → Client:** `game_state`, `chat`, `player_joined`, `game_over`, `error`

---

## **Adding a New Game**

To add a new game to BACOPA, you only need to:

1. **Create `rlgb/games/your_game.py`** — implement the `Game` interface
2. **Register it in `rlgb/serialization.py`** — add to `GAME_REGISTRY` with display conversion
3. **Add a frontend component** — `src/components/YourGame.tsx` for rendering

That's it. The Go backend, WebSocket handler, AI, lobby, and matchmaking all work automatically.

---

## **License: AGPLv3**

BACOPA is licensed under the **GNU Affero General Public License v3.0**. The AGPLv3 ensures that if anyone takes the BACOPA codebase, modifies it, and hosts it as a web service, they must release their modifications back to the open-source community.
