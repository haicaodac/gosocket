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
    gosocketListen.close = function () {
        gosocketServer.close();
    }

    var connectSocket = function (url) {
        if (!url) {
            throw new Error("URL not found!");
        }
        gosocketServer = new WebSocket(url);
        // gosocketServer.onopen = function () {
        //     console.log("OPEN");
        // }
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
        };
        gosocketServer.onerror = function (event) {
            var func = gosocketEvents["onerror"];
            if (func) {
                func(event)
            }
            gosocketServer.close();
        };
        gosocketServer.onclose = function (event) {
            var func = gosocketEvents["onclose"];
            if (func) {
                func(event)
            }
            setTimeout(function () {
                window.gosocket = connectSocket(url)
            }, 1000);
        };

        return gosocketListen;
    }

    window.gosocket = connectSocket
})(window.gosocketListen);
var gosocket = window.gosocket;