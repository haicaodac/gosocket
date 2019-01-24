(function (gosocketListen) {
    var gosocketServer = null;
    gosocketEvents = [];

    gosocketListen = {};
    gosocketListen.on = function (type, func) {
        gosocketEvents[type] = func;
    }
    gosocketListen.emit = function (type, data) {
        function isObject(o) {
            return o !== null && typeof o === 'object' && Array.isArray(o) === false;
        }
        if (!isObject(data)) {
            new Error("The data must be json");
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
            return new Error("URL not found!");
        }
        gosocketServer = new WebSocket(url);
        gosocketServer.onmessage = function (e) {
            var data = JSON.parse(e.data);
            if (data.type) {
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