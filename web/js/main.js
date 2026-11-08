let currentUser = null;

async function api(method, url, body) {
    const options = { method: method, headers: {} };
    if (body) {
        options.headers["Content-Type"] = "application/json";
        options.body = JSON.stringify(body);
    }
    const res = await fetch(url, options);
    const data = await res.json().catch(() => null);
    if (!res.ok) {
        throw new Error(data && data.error ? data.error : "request failed");
    }
    return data;
}

function $(id) {
    return document.getElementById(id);
}

function show(el, visible) {
    el.classList.toggle("hidden", !visible);
}

function setMessage(id, text, isError) {
    const el = $(id);
    el.textContent = text;
    el.classList.toggle("error", !!isError);
}

function formData(form) {
    return Object.fromEntries(new FormData(form));
}

function makePostCard(post, extra) {
    const card = document.createElement("article");
    card.className = "card post";

    const title = document.createElement("h4");
    title.textContent = post.Title || "(no title)";
    card.appendChild(title);

    const meta = document.createElement("div");
    meta.className = "meta";
    meta.textContent = (post.AuthorUsername || "unknown") + " · " + new Date(post.CreatedAt).toLocaleString();
    card.appendChild(meta);

    if (post.Content) {
        const text = document.createElement("p");
        text.textContent = post.Content;
        card.appendChild(text);
    }

    if (post.MediaURL && post.MediaURL.startsWith("http")) {
        const link = document.createElement("a");
        link.href = post.MediaURL;
        link.target = "_blank";
        link.rel = "noopener";
        link.textContent = post.MediaURL;
        card.appendChild(link);
    }

    if (extra) {
        card.appendChild(extra);
    }
    return card;
}

function fillSelect(select, values, selected) {
    select.innerHTML = "";
    for (const v of values) {
        const opt = document.createElement("option");
        opt.value = v;
        opt.textContent = v;
        if (v === selected) {
            opt.selected = true;
        }
        select.appendChild(opt);
    }
}

// Page shell

function openPage(name) {
    for (const btn of document.querySelectorAll(".nav-btn")) {
        btn.classList.toggle("active", btn.dataset.page === name);
    }
    for (const page of document.querySelectorAll(".page")) {
        show(page, page.id === "page-" + name);
    }
}

function showApp() {
    show($("auth-screen"), false);
    show($("topbar"), true);
    show($("app"), true);
    $("current-user").textContent = currentUser.username + " (" + currentUser.role + ")";
    show($("nav-moderation"), currentUser.role === "moderator");
    openPage("feed");
    loadExternalOptions();
    if (currentUser.role === "moderator") {
        loadStrategies();
    }
}

function showAuth() {
    currentUser = null;
    show($("topbar"), false);
    show($("app"), false);
    show($("auth-screen"), true);
}

for (const btn of document.querySelectorAll(".nav-btn")) {
    btn.addEventListener("click", () => openPage(btn.dataset.page));
}

$("logout-btn").addEventListener("click", async () => {
    await api("POST", "/api/logout").catch(() => null);
    showAuth();
});

// Login screen

for (const tab of document.querySelectorAll(".tab")) {
    tab.addEventListener("click", () => {
        for (const t of document.querySelectorAll(".tab")) {
            t.classList.toggle("active", t === tab);
        }
        show($("login-form"), tab.dataset.tab === "login");
        show($("register-form"), tab.dataset.tab === "register");
        setMessage("auth-message", "");
    });
}

$("login-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    try {
        currentUser = await api("POST", "/api/login", formData(e.target));
        e.target.reset();
        setMessage("auth-message", "");
        showApp();
    } catch (err) {
        setMessage("auth-message", err.message, true);
    }
});

$("register-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = formData(e.target);
    try {
        await api("POST", "/api/register", data);
        currentUser = await api("POST", "/api/login", { login: data.email, password: data.password });
        e.target.reset();
        setMessage("auth-message", "");
        showApp();
    } catch (err) {
        setMessage("auth-message", err.message, true);
    }
});

// External content

async function loadExternalOptions() {
    try {
        const sources = await api("GET", "/api/external/sources");
        fillSelect($("external-source"), sources.sort());
        fillSelect($("external-strategy"), ["strict", "banned_words", "caps_spam", "blocked_links", "none"], "strict");
    } catch (err) {
        setMessage("external-message", err.message, true);
    }
}

$("external-load").addEventListener("click", async () => {
    const list = $("external-list");
    list.innerHTML = "";
    setMessage("external-message", "Loading...");
    const source = encodeURIComponent($("external-source").value);
    const strategy = encodeURIComponent($("external-strategy").value);
    try {
        const posts = await api("GET", "/api/external?source=" + source + "&strategy=" + strategy);
        setMessage("external-message", posts.length + " posts passed moderation");
        for (const p of posts) {
            list.appendChild(makePostCard(p));
        }
    } catch (err) {
        setMessage("external-message", err.message, true);
    }
});

// Moderation panel

async function loadStrategies() {
    try {
        const names = await api("GET", "/api/moderation/strategies");
        fillSelect($("mod-strategy"), names, "strict");
    } catch (err) {
        setMessage("mod-message", err.message, true);
    }
}

async function loadModerationPosts() {
    const list = $("mod-list");
    list.innerHTML = "";
    const strategy = encodeURIComponent($("mod-strategy").value);
    const onlyFlagged = $("mod-only-flagged").checked;
    try {
        const items = await api("GET", "/api/moderation/posts?strategy=" + strategy);
        const shown = items.filter((it) => !onlyFlagged || !it.result.allowed);
        const flagged = items.filter((it) => !it.result.allowed).length;
        setMessage("mod-message", items.length + " posts checked, " + flagged + " flagged");
        for (const it of shown) {
            list.appendChild(makePostCard(it.post, moderationControls(it)));
        }
    } catch (err) {
        setMessage("mod-message", err.message, true);
    }
}

function moderationControls(item) {
    const box = document.createElement("div");
    box.className = "mod-controls";

    const badge = document.createElement("span");
    badge.className = item.result.allowed ? "badge ok" : "badge bad";
    badge.textContent = item.result.allowed ? "OK" : "Flagged: " + item.result.reason;
    box.appendChild(badge);

    const del = document.createElement("button");
    del.className = "btn-small danger";
    del.textContent = "Delete";
    del.addEventListener("click", async () => {
        if (!confirm("Delete this post?")) {
            return;
        }
        try {
            await api("DELETE", "/api/moderation/posts/" + item.post.ID);
            loadModerationPosts();
        } catch (err) {
            setMessage("mod-message", err.message, true);
        }
    });
    box.appendChild(del);
    return box;
}

$("mod-load").addEventListener("click", loadModerationPosts);

$("mod-check-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = formData(e.target);
    data.strategy = $("mod-strategy").value;
    try {
        const res = await api("POST", "/api/moderation/check", data);
        if (res.allowed) {
            setMessage("mod-check-result", "Allowed by " + data.strategy);
        } else {
            setMessage("mod-check-result", "Blocked: " + res.reason, true);
        }
    } catch (err) {
        setMessage("mod-check-result", err.message, true);
    }
});

// Start

api("GET", "/api/me")
    .then((user) => {
        currentUser = user;
        showApp();
    })
    .catch(showAuth);
