window.initInputRange = function(r) {
    var s = document.createElement('style'), 
        prefs = ['webkit-slider-runnable', 'moz-range'];

    document.body.appendChild(s);

    var getTrackStyleStr = function(el, val) {
        var str = '', 
            len = prefs.length;

        for(var i = 0; i < len; i++) {
            str += '.js input[type=range]:focus::-' + prefs[i] + '-track{background-size:' + val + '}';
        }

        return str;
    };

    var getTipStyleStr = function(el, val) {
        // /deep/ doesn't work anymore but whatever...
        var str = '.js input[type=range]:focus /deep/ #thumb:after{content:"' + 
            val + '"}';

        return str;
    };

    var getValStr = function(el, p) {
        var min = el.min || 0, 
        perc = (el.max) ? ~~(100*(p - min)/(el.max - min)) : p, 
        val = '20% 100%, ' + perc + '% 100%';

        return val;
    };

    r.addEventListener('input', function() {
        s.textContent = getTrackStyleStr(
                this, 
                getValStr(this, this.value)
                );
        s.textContent += getTipStyleStr(this, this.value);
    }, false);
};
