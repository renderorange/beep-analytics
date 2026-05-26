(function() {
    var r = new XMLHttpRequest();
    r.open('POST', '/collect', true);
    r.setRequestHeader('Content-Type', 'application/json');
    r.send(JSON.stringify({
        origin: location.origin,
        path: location.pathname,
        referrer: document.referrer,
        screen: screen.width + 'x' + screen.height
    }));
})();
