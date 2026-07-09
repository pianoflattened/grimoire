if (typeof window.__grimoirelistener !== "undefined") {
    try {browser.tabs.onUpdated.removeListener(window.__grimoirelistener);}
    catch (e) {}
}

const listener = async (tabId, changeInfo, tab) => {
    if (changeInfo.status !== "loading") return;
    if (!tab.url.startsWith("http://") && !tab.url.startsWith("https://")) return;
    port.postMessage(`NEWPAGE;${btoa(tab.url)};${tabId}`);
}

browser.tabs.onUpdated.addListener(listener);
window.__grimoirelistener = listener;

aj = function(code, tabId) {
    browser.tabs.executeScript(tabId, {code: code});
    return "OK";
}; window.aj = aj;

ac = function(code, tabId) {
    browser.tabs.insertCSS(tabId, {code: code});
    return "OK";
}; window.ac = ac;

"OK";
