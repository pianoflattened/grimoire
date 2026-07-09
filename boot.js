if (typeof window.rec === "undefined") {window.rec = {}};
if (typeof window.gl !== "undefined") {
    try {browser.tabs.onUpdated.removeListener(window.gl);}
    catch (e) {}
}

const listener = async (tabId, changeInfo, tab) => {
    if (changeInfo.status !== "loading") return;
    if (!tab.url.startsWith("http://") && !tab.url.startsWith("https://")) return;
    window.rec[tabId] = {};
    port.postMessage(`NEWPAGE;${btoa(tab.url)};${tabId}`);
}

browser.tabs.onUpdated.addListener(listener);
window.gl = listener;

j = function(code, tabId) {
    const hash = btoa(code);
    if (typeof window.rec[tabId][hash] === "undefined") {
        window.rec[tabId][hash] = true;
        browser.tabs.executeScript(tabId, {
            code: code,
            runAt: "document_start"
        });        
    }
    return "OK";
}; window.j = j;

c = function(code, tabId) {
    const hash = btoa(code);
    if (typeof window.rec[tabId][hash] === "undefined") {
        browser.tabs.insertCSS(tabId, {
            code: code,
            runAt: "document_start"
        });        
    }
    return "OK";
}; window.c = c;

"OK";
