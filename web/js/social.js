
let bellTimer = null;

function initSocial() {
    refreshBell();
    clearInterval(bellTimer);
    bellTimer = setInterval(refreshBell, 30000);
}

function stopSocial() {
    clearInterval(bellTimer);
    bellTimer = null;
    show($("bell-menu"), false);
    show($("bell-count"), false);
}

async function refreshBell() {
    try {
        const res = await api("GET", "/api/notifications/unread-count");
        $("bell-count").textContent = res.count;
        show($("bell-count"), res.count > 0);
    } catch (err) {
        show($("bell-count"), false);
    }
}

function notificationText(n) {
    let what = "commented on your post";
    if (n.Type === "reaction") {
        what = n.CommentID ? "reacted to your comment" : "reacted to your post";
    }
    return n.ActorUsername + " " + what;
}

function makeNotificationRow(n) {
    const row = document.createElement("div");
    row.className = n.ReadAt ? "notif" : "notif unread";

    const text = document.createElement("div");
    text.textContent = notificationText(n);
    row.appendChild(text);

    const when = document.createElement("div");
    when.className = "meta";
    when.textContent = new Date(n.CreatedAt).toLocaleString();
    row.appendChild(when);

    row.addEventListener("click", async () => {
        if (!row.classList.contains("unread")) {
            return;
        }
        try {
            await api("POST", "/api/notifications/" + n.ID + "/read");
            row.classList.remove("unread");
            refreshBell();
        } catch (err) {
            when.textContent = err.message;
        }
    });
    return row;
}

async function openBellMenu() {
    const menu = $("bell-menu");
    menu.innerHTML = "";
    try {
        const list = await api("GET", "/api/notifications");
        if (list.length === 0) {
            const empty = document.createElement("div");
            empty.className = "notif-empty";
            empty.textContent = "No notifications yet.";
            menu.appendChild(empty);
        }
        for (const n of list) {
            menu.appendChild(makeNotificationRow(n));
        }
    } catch (err) {
        const bad = document.createElement("div");
        bad.className = "notif-empty";
        bad.textContent = err.message;
        menu.appendChild(bad);
    }
    show(menu, true);
}

$("bell-btn").addEventListener("click", (e) => {
    e.stopPropagation();
    if ($("bell-menu").classList.contains("hidden")) {
        openBellMenu();
    } else {
        show($("bell-menu"), false);
    }
});

document.addEventListener("click", (e) => {
    if (!$("bell-menu").contains(e.target)) {
        show($("bell-menu"), false);
    }
});

function makeReactions(summary, url) {
    let state = {
        LikeCount: summary.LikeCount || 0,
        DislikeCount: summary.DislikeCount || 0,
        UserReaction: summary.UserReaction || 0,
    };

    const box = document.createElement("div");
    box.className = "reactions";

    const up = document.createElement("button");
    up.className = "react-btn";
    const down = document.createElement("button");
    down.className = "react-btn";
    const note = document.createElement("span");
    note.className = "message error";

    function paint() {
        up.textContent = "▲ " + state.LikeCount;
        down.textContent = "▼ " + state.DislikeCount;
        up.classList.toggle("active", state.UserReaction === 1);
        down.classList.toggle("active", state.UserReaction === -1);
    }

    async function send(value) {
        const next = state.UserReaction === value ? 0 : value;
        try {
            state = await api("POST", url, { value: next });
            note.textContent = "";
            paint();
        } catch (err) {
            note.textContent = err.message;
        }
    }

    up.addEventListener("click", () => send(1));
    down.addEventListener("click", () => send(-1));

    paint();
    box.appendChild(up);
    box.appendChild(down);
    box.appendChild(note);
    return box;
}

function makeComment(c) {
    const item = document.createElement("div");
    item.className = c.ParentCommentID ? "comment reply" : "comment";

    const meta = document.createElement("div");
    meta.className = "meta";
    meta.textContent = c.AuthorUsername + " · " + new Date(c.CreatedAt).toLocaleString();
    item.appendChild(meta);

    const text = document.createElement("div");
    text.textContent = c.Content;
    item.appendChild(text);

    item.appendChild(makeReactions(c.Reactions, "/api/comments/" + c.ID + "/reactions"));
    return item;
}

function makeComments(post) {
    const box = document.createElement("div");
    box.className = "comments";

    const toggle = document.createElement("button");
    toggle.className = "react-btn";
    toggle.textContent = "Comments";
    box.appendChild(toggle);

    const body = document.createElement("div");
    body.className = "comments-body hidden";
    box.appendChild(body);

    const list = document.createElement("div");
    body.appendChild(list);

    const form = document.createElement("form");
    form.className = "comment-form";
    const input = document.createElement("textarea");
    input.rows = 2;
    input.maxLength = 1000;
    input.placeholder = "Write a comment";
    input.required = true;
    const send = document.createElement("button");
    send.type = "submit";
    send.textContent = "Comment";
    const note = document.createElement("span");
    note.className = "message error";
    form.appendChild(input);
    form.appendChild(send);
    form.appendChild(note);
    body.appendChild(form);

    async function load() {
        list.innerHTML = "";
        try {
            const comments = await api("GET", "/api/posts/" + post.ID + "/comments");
            toggle.textContent = "Comments (" + comments.length + ")";
            for (const c of comments) {
                list.appendChild(makeComment(c));
            }
        } catch (err) {
            note.textContent = err.message;
        }
    }

    let loaded = false;
    toggle.addEventListener("click", () => {
        show(body, body.classList.contains("hidden"));
        if (!loaded) {
            loaded = true;
            load();
        }
    });

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        try {
            await api("POST", "/api/posts/" + post.ID + "/comments", { content: input.value });
            input.value = "";
            note.textContent = "";
            load();
        } catch (err) {
            note.textContent = err.message;
        }
    });

    return box;
}

function makeSocialBar(post) {
    const bar = document.createElement("div");
    bar.className = "social-bar";
    bar.appendChild(makeReactions(post.Reactions || {}, "/api/posts/" + post.ID + "/reactions"));
    bar.appendChild(makeComments(post));
    return bar;
}
