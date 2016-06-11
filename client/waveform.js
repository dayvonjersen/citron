const pixelRatio = window.devicePixelRatio || screen.deviceXDPI / screen.logicalXDPI;
const minPxPerSec = 20;
const waveColor = "#fefbf6";
const normalize = false;

function processAudioFile(audioFile, doneFn) {
    let reader = new FileReader();
    reader.onload = (event) => _processAudioFile(event.target.result, doneFn);
    reader.readAsArrayBuffer(audioFile);
}

function _processAudioFile(audioData, doneFn) {
    let audioCtx = new AudioContext();
    let canvasElement = document.createElement('canvas');
    let canvasCtx = canvasElement.getContext('2d');

    audioCtx.decodeAudioData(audioData, (audioBuffer) => {
        let width = Math.round(audioBuffer.duration * minPxPerSec * pixelRatio);
        canvasElement.width = width;
        canvasElement.height = 128;
        let peaks = getPeaks(audioBuffer, width);
        drawPeaks(canvasCtx, peaks, width);
        doneFn(canvasElement.toDataURL(), audioBuffer.duration);
    });
}

function getPeaks(audioBuffer, length, splitChannels=false) {
    let sampleSize = audioBuffer.length / length;
    let sampleStep = ~~(sampleSize / 10) || 1;
    let channels = audioBuffer.numberOfChannels;
    let splitPeaks = [];
    let mergedPeaks = [];

    for (let c = 0; c < channels; c++) {
        let peaks = splitPeaks[c] = [];
        let chan = audioBuffer.getChannelData(c);

        for (let i = 0; i < length; i++) {
            let start = ~~(i * sampleSize);
            let end = ~~(start + sampleSize);
            let min = 0;
            let max = 0;

            for (let j = start; j < end; j += sampleStep) {
                let value = chan[j];

                if (value > max) {
                    max = value;
                }

                if (value < min) {
                    min = value;
                }
            }

            peaks[2 * i] = max;
            peaks[2 * i + 1] = min;

            if (c == 0 || max > mergedPeaks[2 * i]) {
                mergedPeaks[2 * i] = max;
            }

            if (c == 0 || min < mergedPeaks[2 * i + 1]) {
                mergedPeaks[2 * i + 1] = min;
            }
        }
    }

    return splitChannels ? splitPeaks : mergedPeaks;
}

function drawPeaks(canvasCtx, peaks, width, splitChannels=false, channelIndex=0) {
    // Split channels
    /*
       if (peaks[0] instanceof Array) {
       let channels = peaks;
       if (splitChannels) {
       this.setHeight(channels.length * 128 * pixelRatio);
       channels.forEach(this.drawWave, this);
       return;
       } else {
       peaks = channels[0];
       }
       }
       */

    // Support arrays without negative peaks
    let hasMinValues = [].some.call(peaks, (val) => val < 0);
    if (!hasMinValues) {
        let reflectedPeaks = [];
        for (let i = 0, len = peaks.length; i < len; i++) {
            reflectedPeaks[2 * i] = peaks[i];
            reflectedPeaks[2 * i + 1] = -peaks[i];
        }
        peaks = reflectedPeaks;
    }

    // A half-pixel offset makes lines crisp
    let $ = 0.5 / pixelRatio;
    let height = 128;
    let offsetY = height * channelIndex || 0;
    let halfH = height / 2;
    let length = ~~(peaks.length / 2);

    let scale = 1;
    if (width != length) {
        scale = width / length;
    }

    let absmax = 1;
    if (normalize) {
        let max = Math.max.apply(Math, peaks);
        let min = Math.min.apply(Math, peaks);
        absmax = -min > max ? -min : max;
    }

    canvasCtx.fillStyle = waveColor;
    
    canvasCtx.beginPath();
    canvasCtx.moveTo($, halfH + offsetY);

    for (let i = 0; i < length; i++) {
        let h = Math.round(peaks[2 * i] / absmax * halfH);
        canvasCtx.lineTo(i * scale + $, halfH - h + offsetY);
    }

    // Draw the bottom edge going backwards, to make a single
    // closed hull to fill.
    for (let i = length - 1; i >= 0; i--) {
        let h = Math.round(peaks[2 * i + 1] / absmax * halfH);
        canvasCtx.lineTo(i * scale + $, halfH - h + offsetY);
    }

    canvasCtx.closePath();
    canvasCtx.fill();

    // Always draw a median line
    canvasCtx.fillRect(0, halfH + offsetY - $, width, $);
}
