(function (gosocketListen) {
    var gosocketServer = null;
    gosocketEvents = [];

    gosocketListen = {};
    gosocketListen.on = function (type, func) {
        gosocketEvents[type] = func;
    }
    gosocketListen.emit = function (type, data) {
        if (!data || typeof data !== "object" || data.length > 0) {
            throw new Error("The data must be json");
        }
        var message = {
            "type": type,
            "content": data
        }
        message = JSON.stringify(message);
        gosocketServer.send(message)
    }

    window.gosocket = function (url) {
        if (!url) {
            throw new Error("URL not found!");
        }
        gosocketServer = new WebSocket(url);
        gosocketServer.onmessage = function (e) {
            try {
                var data = JSON.parse(e.data);
            } catch (e) {
                console.Error("The data must be json");
            }
            if (data && data.type) {
                var func = gosocketEvents[data.type];
                if (func) {
                    func(data.content)
                }
            }
        }
        return gosocketListen;
    }
})(window.gosocketListen);
var gosocket = window.gosocket;