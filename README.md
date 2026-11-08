# Final Project — Social Media Feed
Project about creating Social Media Feed 

## Authors – Group SE-2534:
- Nuraly Auyelkhan
- Yeren Yedige
- Muratov Anuar

## Structure of the project:
```text
├── cmd/
│   └── server/
│       └── main.go              
├── internal/
│   ├── api/                      # HTTP handlers (auth, moderation, external)
│   ├── auth/                     # Users, sessions, login and register
│   ├── adapter/                  # Pattern Adapter
│   |   └── adapter.go
│   ├── observer/                 # Pattern Observer
│   |   └── observer.go
│   ├── strategy/                 # Pattern Strategy
│   |   └── strategy.go
│   ├── facade/                 # Pattern Facade
│   |   └── facade.go
│   ├── decorator/                 # Pattern Decorator
│   |   └── decorator.go
│   └── factory/                 # Pattern Factory
│      └── factory.go
├── web/                          # Front-end
│   ├── js/
│   |   └── main.js
│   ├── css/
│   |   └── main.css
│   └── index.html     
└── README.md                     # Documentation
```
## Database - SQLite 

SQLite is a powerful database for listing settings and user info.

## Functionality 
- Content creation block
- Post feed
- Social Interaction and Notifications
- Moderation and processing of posts
- Integration of external content

## How to run
```bash
go run ./cmd/server
```
Open http://localhost:8080. The first registered user becomes a moderator.
