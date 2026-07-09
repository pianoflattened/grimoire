if (typeof window.last === "undefined") window.last = {};
if (typeof window.gl !== "undefined") {
    try {browser.tabs.onUpdated.removeListener(window.gl);}
    catch (e) {}
}

const listener = async (tabId, changeInfo, tab) => {
    if (changeInfo.status !== "loading" || !changeInfo.url) return;
    if (!tab.url.startsWith("http://") && !tab.url.startsWith("https://")) return;
    const now = Date.now();
    const last = window.last[tabId];
    if (last && last.url && (now - last.time) < 150) return;
    window.last[tabId] = {url: tab.url, time: now};
    port.postMessage(`NEWPAGE;${btoa(tab.url)};${tabId}`);
}

browser.tabs.onUpdated.addListener(listener);
window.gl = listener;

j = function(code, tabId) {
    browser.tabs.executeScript(tabId, {
        code: code,
        runAt: "document_start"
    });
}; window.j = j;

c = function(code, tabId) {
    browser.tabs.insertCSS(tabId, {
        code: code,
        runAt: "document_start"
    });
}; window.c = c;

"OK";
