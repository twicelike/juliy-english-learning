Juliy English is an AI-powered English vocabulary learning platform
designed to help learners discover, organize, practice, and retain
English vocabulary through meaning-aware learning tools and spaced
repetition.
Project status: In active development. Features and integrations
may vary by the current release.

Overview
Traditional vocabulary tools often treat a word as a single item, even
when it has several meanings. Vocabit is designed around sense-level
vocabulary learning so learners can work with a specific meaning of a
word instead of repeatedly encountering an arbitrary definition.

The platform combines vocabulary discovery and personal word collections
with review workflows. AI-powered capabilities are part of the product
direction, including vocabulary explanations and writing practice.
Key Features
- Dictionary --- Search English vocabulary and explore word
  information, parts of speech, definitions, and examples where
  available.
- Word Sets --- Create and manage personal vocabulary collections
  for focused study.
- Match / Word Discovery --- Discover vocabulary through a
  card-based interface and add useful words to a collection.
- Vocabulary Review --- Practice saved vocabulary with flashcards
  and spaced-repetition workflows.
- Sense-level learning model --- Represent individual meanings of
  polysemous words separately, helping learning progress and exercises
  target the intended meaning.
- AI-assisted learning --- Planned and/or in-progress integrations
  for vocabulary explanations and writing practice, depending on the
  deployed version.
Tech Stack
  Layer                               Technologies
  Frontend                            Next.js 16, React, TypeScript,
                                      Tailwind CSS

  UI & Animation                      Framer Motion

  Backend                             Go (Golang), Gin

  Architecture                        Clean Architecture,
                                      handler/service/domain/repository
                                      separation

  Database                            MongoDB, MongoDB Go Driver v2

  AI providers                        Gemini, Groq, Ollama, OpenRouter
                                      --- provider support depends on
                                      current integration status

  Development / deployment            Windows development environment,
                                      Cloudflare Tunnel for exposing the
                                  development deployment
Only list integrations as production-ready when they are implemented
and configured in the version you are presenting. Redis, Docker
Compose, and other planned infrastructure are intentionally not listed
as current dependencies here.

Architecture
At a high level, the application is organized into the following layers:
Browser
  |
  v
Next.js frontend
  |
  v
Go API (Gin)
  |
  v
Handlers
  |
  v
Services / business logic
  |
  v
Domain models and rules
  |
  v
Repositories
  |
  v
MongoDB

Services ------> AI provider adapter(s) ------> External or local AI models
The backend separates transport concerns, application logic, domain
rules, and persistence to make the codebase easier to maintain and test.
Repository Structure
The exact layout may evolve as development continues. A typical layout
for this project is:
vocabit/
├── frontend/              # Next.js application
└── logic/                 # Go backend
    ├── cmd/                # Application entry points (if present)
    └── internal/           # Feature and application packages
        ├── wordset/
        ├── match/
        └── ...
Adjust these paths to match the actual repository before publishing the
README.
Getting Started
The exact setup commands depend on the current repository structure and
environment configuration. The following is a general local-development
checklist.
Prerequisites
- Go version compatible with the backend's go.mod
- Node.js version compatible with the frontend's package.json
- MongoDB
- Git
1. Clone the repository
git clone <repository-url>
cd <repository-directory>
2. Configure MongoDB
Start MongoDB locally and configure the connection string expected by
the backend.

Example local connection string:
mongodb://localhost:27017/dictionary
Use the database and collection names expected by the current
configuration. Do not commit credentials or production secrets.
3. Run the backend
From the Go backend directory (currently referred to as logic in the
project plan):
go mod download
go run .
If the application entry point is under cmd/, run the command for that
entry point instead. Check the project's current configuration for
required environment variables and the correct port.
4. Run the frontend
From the Next.js frontend directory:
npm install
npm run dev
Open the local URL printed by Next.js, commonly http://localhost:3000.
These commands are a starting point, not a guarantee of the current
repository's exact entry points. Update them to match the checked-in
scripts and configuration.

Engineering Focus
- Meaning-aware vocabulary: Model a word's individual senses
  rather than relying only on a word-level identifier or definition
  position.
- Maintainable backend: Keep HTTP handlers thin and put
  application rules in services and domain logic.
- Review scheduling: Use spaced-repetition concepts to prioritize
  vocabulary that needs practice.
- AI cost and latency: Prefer reusable, stored results where
  appropriate and consider quality control before promoting generated
  content into shared caches.
- Incremental delivery: Build core vocabulary and review workflows
  before expanding more expensive AI features.
Roadmap
The roadmap is subject to change. Areas of ongoing or planned work
include:
- Refining sense-level learning, bookmarks, word sets, and review
  sessions.
- Improving vocabulary discovery and translation/annotation workflows.
- Expanding AI-assisted explanations and writing practice.
- Improving response latency, cache quality, and operational
  reliability.
- Increasing test coverage and documenting deployment procedures.
Security
- Keep API keys and database credentials in environment variables or a
  secret manager.
- Never commit .env files containing secrets.
- Validate and authorize user-specific operations on the backend.
- Use secure, HttpOnly cookies for session-based authentication if
  that authentication flow is enabled in the current build.
- Treat AI-generated explanations and feedback as content that may
  require validation and quality controls.
