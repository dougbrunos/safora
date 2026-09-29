(function () {
  var root = document.documentElement;
  var TITLES = {
    pt: "Safora: transforme scripts de backup em tarefas agendadas",
    en: "Safora: turn backup scripts into scheduled jobs",
  };

  // Language: saved choice, otherwise the browser's.
  function saved() {
    try { return localStorage.getItem("safora_site_lang"); } catch (e) { return null; }
  }
  function setLang(lang) {
    root.lang = lang;
    document.title = TITLES[lang];
    try { localStorage.setItem("safora_site_lang", lang); } catch (e) { /* not persisted */ }
  }
  var param = new URLSearchParams(location.search).get("lang");
  var initial = (param === "pt" || param === "en") ? param : saved() || ((navigator.language || "pt").toLowerCase().indexOf("pt") === 0 ? "pt" : "en");
  setLang(initial);
  document.getElementById("lang-toggle").addEventListener("click", function () {
    setLang(root.lang === "pt" ? "en" : "pt");
  });

  // Download buttons point to the newest release; the static links stay as a fallback.
  var RELEASES = "https://api.github.com/repos/dougbrunos/safora/releases?per_page=5";
  var MATCH = {
    "windows": /^Safora-Setup-.*\.exe$/,
    "linux-amd64": /-linux-amd64\.tar\.gz$/,
    "linux-arm64": /-linux-arm64\.tar\.gz$/,
    "sums": /^SHA256SUMS\.txt$/,
  };
  fetch(RELEASES, { headers: { Accept: "application/vnd.github+json" } })
    .then(function (r) { return r.ok ? r.json() : Promise.reject(); })
    .then(function (list) {
      var release = list.filter(function (r) { return !r.draft; })[0];
      if (!release) return;
      document.querySelectorAll("[data-asset]").forEach(function (a) {
        var asset = release.assets.filter(function (x) { return MATCH[a.dataset.asset].test(x.name); })[0];
        if (asset) a.href = asset.browser_download_url;
      });
      var ver = document.getElementById("ver");
      ver.textContent = release.tag_name;
      ver.hidden = false;
    })
    .catch(function () { /* keep the fallback links */ });
})();
