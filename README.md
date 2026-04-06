# **BACOPA**

**Open Creative Multi-Game Playground Where Humans, Creators, Players, And Bots Meet**

## **What is BACOPA?**

*Bacopa monnieri* is an aquatic plant known for two things: its rapid, dense branching structure, and its historical use as a nootropic (cognitive enhancer).  
**BACOPA** the platform is exactly that:

1. **Rapid Branching:** A playground where game designers and enthusiasts can endlessly "fork" existing games, tweak rulesets, and invent entirely new variants.  
2. **Cognitive Enhancement:** Powered by the [**RLGB (Reinforcement Learning Game Bot)**](https://github.com/horkah/RLGB) engine, BACOPA automatically trains state-of-the-art AI agents to play these user-generated games using Proximal Policy Optimization (PPO), self-play, and evolutionary curriculum learning.

BACOPA bridges the gap between traditional gaming, tabletop design, and modern machine learning. It is a unified ecosystem where players can face off against human or bot opponents, and designers can instantly test their chaotic new mechanics against intelligent agents.

## **The Technology Stack**

Given the ambitious scope of bridging high-performance machine learning with real-time web interfaces, BACOPA utilizes a decoupled microservice architecture.  
| **Component** | **Technology** | **Why it's the Best Choice** |  
| **Frontend Framework** | React (Next.js) \+ TypeScript | Next.js handles fast initial loads and complex routing. TypeScript is absolutely mandatory to maintain sanity with unpredictable, dynamic game state objects across a large team. |  
| **Game Renderer** | PixiJS (2D) / Three.js (3D) | Lightweight, headless-compatible renderers that can dynamically draw boards based on abstract backend JSON states. |  
| **Platform Backend** | Node.js (NestJS) or Go | High concurrency. Manages matchmaking, user profiles, Elo ratings, WebSocket routing, and metadata handling. |  
| **Live Inference API** | Python (FastAPI) | Wraps the RLGB engine. Asynchronous, microsecond-latency API to feed game states to ML models and return AI moves instantly. |  
| **Real-time Comms** | WebSockets (Socket.io) | Essential for real-time multiplayer moves, chat, and live game observation. |  
| **Relational DB** | PostgreSQL | Robust handling of users, complex ruleset schemas, and the intricate "forking" histories of custom games. |  
| **In-Memory DB** | Redis | Crucial for matchmaking queues, holding active live game session states, and message brokering. |  
| **MLOps & Infra** | Docker, Kubernetes, Celery | Orchestrates the auto-scaling of GPU pods required to train new agents when a user publishes a new game. |

## **Core Architecture & Design Requirements**

BACOPA's architecture is split into three highly distinct domains.

### **1\. The Live Arena (State Management & Inference)**

RLGB's core design dictates that games are stateless and transition functionally (next\_state(state, action)).

* The **Platform Backend** (Node/Go) holds the "ground truth" of any active multiplayer game in Redis.  
* When it is a Bot's turn, the Platform Backend sends the current state via gRPC/HTTP to the **Inference API** (FastAPI).  
* FastAPI dynamically loads the corresponding PyTorch model (weights.pt & metadata.json) for that specific game variant, runs agent.act(), and instantly returns the move.

### **2\. The Universal Compiler (The Grand Challenge)**

Users create games in a visual drag-and-drop web editor. RLGB, however, expects a strict Python Game subclass with explicit mathematical action spaces and perspective-relative observe(state) tensors.

* BACOPA must include a **Compiler Layer**.  
* This layer accepts a universal JSON/YAML schema from the frontend editor and automatically translates it into a valid, immutable RLGB Python class on the fly, bridging visual mechanics with neural network tensor inputs.

### **3\. Asynchronous MLOps (The Training Pipeline)**

When a user invents or forks a game and clicks "Publish":

* A message is sent to a job queue (Celery/RabbitMQ).  
* Kubernetes spins up a GPU training pod.  
* The pod executes RLGB's run\_optimize.py, utilizing curriculum learning and league opponent pools to train a baseline bot from scratch via self-play.  
* Once training yields a competent agent, the model is saved to cloud storage, and the pod spins down. The new game is now "Live" with an AI opponent ready to play.

## **Instructions for the Development Team**

To prevent chaos among a large programming team working on such a multi-faceted platform, development is divided into four focused "Squads."

### **A. Squad Responsibilities**

1. **The Platform Squad (Node/Go & DB):** Focuses on the ecosystem. User authentication, Elo rating algorithms, WebSocket routing, game lobby state management, and the PostgreSQL schema for tracking the lineage of forked games.  
2. **The Compiler Squad (Python):** Handles the hardest translation task. They design the universal JSON Ruleset Schema and build the parser that turns frontend JSON into robust RLGB Game classes.  
3. **The MLOps & Inference Squad (Python):** Focuses on speed and scale. They wrap RLGB into the ultra-fast FastAPI service for live play and orchestrate the automated GPU training queues in Kubernetes.  
4. **The Interface & Editor Squad (React & WebGL):** Builds the player dashboards, community features, and the visual drag-and-drop game editor.

### **B. Development Guidelines**

* **API-First Development:** Before any squad writes logic, OpenAPI/Swagger specs defining exactly how "Game States" and "Ruleset JSONs" are structured must be negotiated and locked.  
* **Strict Typing & Validation:** Use TypeScript strictly on the frontend/platform backend, and Pydantic models in Python. Dynamic game engines fail catastrophically if data shapes are loose.  
* **Fuzz Testing:** Because users can *invent* games, the platform is vulnerable to infinite loops or contradictory mechanics. The Compiler Squad must implement intense fuzz testing—throwing random, garbage JSON rulesets at the compiler to ensure it gracefully rejects them without crashing backend services.  
* **Trunk-Based Development:** Push small, frequent updates to main hidden behind Feature Flags to avoid massive merge conflicts across the microservices.

## **License Choice: AGPLv3**

BACOPA is licensed under the **GNU Affero General Public License v3.0**.  
Because this is an open platform designed for web deployment, the AGPLv3 closes the "Application Service Provider" loophole. It ensures that if any large entity takes the BACOPA codebase, modifies it, and hosts it as a competing web service, they *must* release their modifications back to the open-source community. This guarantees BACOPA remains a truly open ecosystem.

## **Getting Started (Local Development)**

### **Prerequisites**

* Docker & Docker Compose (or minikube for Kubernetes simulation)  
* Node.js (v18+)  
* Python (v3.10+)

### **Setup**

1. Clone the repository: git clone https://github.com/your-org/bacopa.git  
2. Initialize the RLGB engine submodule: git submodule update \--init \--recursive  
3. Spin up the infrastructure (PostgreSQL, Redis, RabbitMQ): docker-compose up \-d  
4. Start the Inference API: cd backend-inference && uvicorn main:app \--reload  
5. Start the Platform Backend: cd backend-platform && npm run start:dev  
6. Start the Frontend Client: cd frontend && npm run dev