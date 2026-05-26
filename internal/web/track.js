(function() {
    var scripts = document.getElementsByTagName('script');
    var src = scripts[scripts.length - 1].src;
    var base = src.replace(/\/track\.js$/, '');

    var r = new XMLHttpRequest();
    r.open('POST', base + '/collect', true);
    r.setRequestHeader('Content-Type', 'application/json');
    r.send(JSON.stringify({
        origin: location.origin,
        path: location.pathname,
        referrer: document.referrer,
        screen: screen.width + 'x' + screen.height
    }));
})();
