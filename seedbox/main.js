"use strict";
const WebSocketClient = require('websocket').client;
const WebTorrent = require('webtorrent-hybrid');
const MagnetURI  = require('magnet-uri');
const fs         = require('fs');

const DOWNLOAD_DIR = './downloads/';
const TORRENT_DIR = './torrents/';

let tClient = new WebTorrent();

// seed existing
fs.readdir(TORRENT_DIR, (err, ls) => {
    if(err) return errorMessage(err);
    ls.forEach((torrentFile) => {
        let torrentPath = TORRENT_DIR + torrentFile;
        console.log("adding "+torrentPath);
        let torrentBuf = fs.readFileSync(torrentPath);
        tClient.add(torrentBuf, {
            path: DOWNLOAD_DIR,
        });
    });
});

function addTorrent(magnet) {
    // console.log("would add "+magnet); return;
    let infoHash = MagnetURI.decode(magnet).infoHash;
    if(tClient.get(infoHash)) {
        console.log("already have "+infoHash);
        return;
    }
    tClient.add(magnet, {path: DOWNLOAD_DIR}, (torrent) => {
        let torrentPath = TORRENT_DIR + torrent.infoHash + ".torrent";
        console.log("writing "+torrentPath);
        fs.writeFile(torrentPath, torrent.torrentFile, (err) => {
            if(err) return errorMessage(err);
            console.log("wrote: "+torrentPath);
        });
    });
}        

function errorMessage(err) {
    console.error(err.toString());
}

let client = new WebSocketClient({
    tlsOptions: {
        rejectUnauthorized: false,
        requestCert: true,
        agent: false,
    }
});

let isLoading = true;
let startIndex = 0;
let reachedEnd = false;

client.on('connectFailed', errorMessage);
client.on('connect', (conn) => {
    console.log('connected');
    let sendPayload = (event,message) => {
        let payload = JSON.stringify({event:event,message:message});
        console.log(payload);
        conn.send(payload);
    };
    conn.on('error', errorMessage);
    conn.on('close', () => console.log('disconnected'));
    conn.on('message', (evt) => {
        let data;
        try {
            data = JSON.parse(evt.utf8Data);
            if(!data.event && !data.message) throw new Error();
        } catch(e) {
            errorMessage("Malformed JSON response:"+JSON.stringify(evt.utf8Data));
        }
        if(!data) return;

        console.log("event: "+data.event);
        switch(data.event) {
            case "suprême":
            case "infinitescroll":
                console.log("got: "+data.message.fileName);
                addTorrent(data.message.magnetURI);
                break;
            case "loaded":
                isLoading = false;
                console.log("loaded:"+data.message);
                startIndex += data.message|0;
                if(data.message == 0) {
                    reachedEnd = true;
                } else {
                    sendPayload("range", {start: startIndex+1, limit: 100});
                }
                break;
            case "error":
                errorMessage(data.message);
                break;
        }
    });
});

client.connect('wss://localhost:12345');
