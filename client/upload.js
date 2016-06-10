const parseTorrent = require('parse-torrent');
const createTorrent = require('create-torrent');

let options = {};
createTorrent(infile, options, (err, data) => {
    if(err) {
        console.error(err.stack);
    } else {
        return parseTorrent.toMagnetURI(parseTorrent( data ));
    }
});

