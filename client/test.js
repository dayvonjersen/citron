const WebTorrent   = require('webtorrent');
const dragDrop     = require('drag-drop');

let client = new WebTorrent();

dragDrop('body', (files) => {
    client.seed(files, (torrent) => {
        client.add(torrent.magnetURI, (torrent) => {
            torrent.files[0].appendTo('body');
        });
    });
});
