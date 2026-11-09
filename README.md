# Final Project — Social Media Feed
The project is implemented in Go using the standard library for HTTP routing and vanilla JavaScript on the client side, forming a lightweight Single Page Application.

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
│   ├── adapter/                  # Pattern Adapter
│   │   └── adapter.go
│   ├── api/                      # HTTP handlers
│   │   ├── api.go
│   │   ├── auth.go
│   │   ├── external.go
│   │   ├── moderation.go
│   │   └── posts.go
│   ├── auth/                     # Users, sessions logic
│   │   ├── repository.go
│   │   └── service.go
│   ├── decorator/                # Pattern Decorator
│   │   └── decorator.go
│   ├── facade/                   # Pattern Facade
│   │   ├── facade.go
│   │   └── facade_feed.go
│   ├── factory/                  # Pattern Factory
│   │   └── factory.go
│   ├── observer/                 # Pattern Observer
│   │   ├── notif_test.go
│   │   ├── notificationRepository.go
│   │   └── observer.go
│   ├── store/                    # Database operations (SQLite)
│   │   ├── db.go
│   │   └── post_store.go
│   ├── strategy/                 # Pattern Strategy
│   │   └── strategy.go
│   ├── structs/                  # Data structures / Domain structss
│   │   └── struct.go
│   └── schema.sql                # SQL миграции / создание таблиц
├── web/                          # Front-end
│   ├── css/
│   │   ├── main.css
│   │   └── social.css
│   ├── js/
│   │   ├── main.js
│   │   └── social.js
│   │              
│   └── index.html     
├── go.mod                        # Go module file
├── app.db             # File of the SQLite Database
└── README.md                     # Documentation
```

## UML Diagram:

### Design patterns (all patterns included)

```mermaid
classDiagram
    direction LR

    class Server {
        +Routes(mux)
        +CurrentUser(r) User
    }

    %% Facade
    class Facade {
        +CreatePost(...) PostView
        +Feed(...) PostView[]
        +DeletePost(id)
    }

    %% Factory
    class Factory {
        +Create(kind, title, userID, content, mediaURL) Post
        +Validate(p) error
    }

    %% Strategy
    class ModerationStrategy {
        <<interface>>
        +Name() string
        +Check(p Post) Result
    }
    class Moderator {
        +SetStrategy(s)
        +Check(p Post) Result
        +Filter(posts) Post[]
    }
    class AllowAll
    class BannedWords
    class CapsSpam
    class BlockedLinks
    class Strict

    %% Decorator
    class PostComponent {
        <<interface>>
        +Post() Post
        +Badges() string[]
    }
    class basePost
    class edited
    class flagged
    class PostView {
        +Post
        +Badges string[]
    }

    %% Observer
    class Observer {
        <<interface>>
        +Notify(e Event) error
    }
    class Subject {
        +Subscribe(o Observer)
        +Publish(e Event)
    }
    class Notifier {
        +Notify(e Event) error
    }
    class notificationStore {
        <<interface>>
        +InsertNotification(n)
        +PostAuthorID(id) int
        +CommentAuthorID(id) int
    }
    class Repo {
        +SavePostReactions(...)
        +SaveCommentReactions(...)
        +CreateComment(...)
        +ListComments(...)
        +ListNotificationForUser(id)
    }

    %% Adapter
    class ContentSource {
        <<interface>>
        +Name() string
        +FetchPosts(limit) Post[]
    }
    class JSONPlaceholderAdapter
    class RSSAdapter

    class PostStore {
        +Create(p)
        +GetByID(id, viewerID)
        +List(viewerID, limit, offset)
    }

    Server --> Facade
    Server --> Subject
    Server --> Repo
    Server --> ContentSource

    Facade --> Factory : builds posts
    Facade --> PostStore : saves and loads
    Facade --> Moderator : flag check
    Facade ..> PostView : returns

    ModerationStrategy <|.. AllowAll
    ModerationStrategy <|.. BannedWords
    ModerationStrategy <|.. CapsSpam
    ModerationStrategy <|.. BlockedLinks
    ModerationStrategy <|.. Strict
    Strict o-- ModerationStrategy : chains
    Moderator o-- ModerationStrategy : uses one

    PostComponent <|.. basePost
    PostComponent <|.. edited
    PostComponent <|.. flagged
    edited o-- PostComponent : wraps
    flagged o-- PostComponent : wraps
    PostView ..> PostComponent : built from

    Observer <|.. Notifier
    Subject o-- Observer : subscribers
    Notifier --> notificationStore
    notificationStore <|.. Repo

    ContentSource <|.. JSONPlaceholderAdapter
    ContentSource <|.. RSSAdapter
```

### Data structures (backend part)

```mermaid
classDiagram
    direction LR

    class User {
        +int ID
        +string Email
        +string Username
        +Role Role
        +string PasswordHash
        +time CreatedAt
    }
    class Session {
        +string ID
        +int UserID
        +time ExpiresAt
        +time CreatedAt
    }
    class Post {
        +int ID
        +int UserID
        +string AuthorUsername
        +PostKind Kind
        +string Title
        +string Content
        +string MediaURL
        +ReactionSummary Reactions
        +time CreatedAt
        +time UpdatedAt
    }
    class ReactionSummary {
        +int LikeCount
        +int DislikeCount
        +Reaction UserReaction
    }
    class PostReaction {
        +int UserID
        +int PostID
        +Reaction Value
        +time CreatedAt
        +time UpdatedAt
    }
    class Comment {
        +int ID
        +int PostID
        +int UserID
        +int ParentCommentID
        +string AuthorUsername
        +string Content
        +ReactionSummary Reactions
        +time CreatedAt
        +time UpdatedAt
    }
    class CommentReaction {
        +int UserID
        +int CommentID
        +Reaction Value
        +time CreatedAt
        +time UpdatedAt
    }
    class Notification {
        +int ID
        +int UserID
        +int ActorID
        +string ActorUsername
        +int PostID
        +int CommentID
        +NotificationType Type
        +time ReadAt
        +time CreatedAt
    }
    class Reaction {
        <<enumeration>>
        Like = 1
        None = 0
        Dislike = -1
    }
    class Role {
        <<enumeration>>
        user
        moderator
    }
    class PostKind {
        <<enumeration>>
        text
        image
        link
    }
    class NotificationType {
        <<enumeration>>
        comment
        reaction
    }

    User "1" --> "*" Session : logs in
    User "1" --> "*" Post : writes
    User "1" --> "*" Comment : writes
    Post "1" --> "*" Comment : has
    Comment "0..1" <-- "*" Comment : replies to
    Post "1" --> "*" PostReaction : receives
    Comment "1" --> "*" CommentReaction : receives
    User "1" --> "*" PostReaction : casts
    User "1" --> "*" CommentReaction : casts
    User "1" --> "*" Notification : receives
    Notification --> User : actor
    Notification --> Post : about
    Notification --> Comment : about
    Post *-- ReactionSummary
    Comment *-- ReactionSummary
    User --> Role
    Post --> PostKind
    PostReaction --> Reaction
    CommentReaction --> Reaction
    Notification --> NotificationType
```

## 6 design patterns:
Adapter: Acts as a "translator" for external sources (JSONPlaceholder API, RSS feeds). Downloads data in other formats (JSON, XML) and brings them to a single ContentSource domain interface and the internal structure of the Post.

Observer: Implements a notification system. Monitors events within the system (new comment, reaction) and sends alerts to target users without tightly linking components.

Strategy: Controls the logic of the moderation panel. It allows you to switch text verification algorithms on the fly (strict verification, search for forbidden words, link blocking) without changing the main handler code.

Factory / Builder: Responsible for encapsulating the logic of creating posts. Performs strict data validation (length limits in UTF-8 runes, content type checking) before the object gets into the database layer.

Facade: Hides the complex logic of feed aggregation. It provides controllers with a simple interface for receiving posts, collecting data from the database and external adapters in a single stream under the hood.

Decorator: Dynamically expands the basic functionality of services (for example, by adding logging, access control, or caching) without interfering with their original code.

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

## How to run tests

```bash
go test ./...
```