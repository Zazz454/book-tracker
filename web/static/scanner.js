// Barcode scanner using BarcodeDetector API (Chrome Android native)
// Falls back gracefully if not supported.

let scannerStream = null;
let scannerInterval = null;

async function startScanner() {
    const container = document.getElementById('scannerContainer');
    const video = document.getElementById('scannerVideo');
    const startBtn = document.getElementById('startScan');
    const stopBtn = document.getElementById('stopScan');

    if (!('BarcodeDetector' in window)) {
        alert('Barcode scanning is not supported in this browser. Please use Chrome on Android, or enter the ISBN manually below.');
        return;
    }

    try {
        scannerStream = await navigator.mediaDevices.getUserMedia({
            video: { facingMode: 'environment' }
        });
        video.srcObject = scannerStream;
        container.classList.add('active');
        startBtn.style.display = 'none';
        stopBtn.style.display = 'inline-flex';

        const detector = new BarcodeDetector({
            formats: ['ean_13', 'ean_8', 'upc_a', 'upc_e']
        });

        scannerInterval = setInterval(async () => {
            try {
                const barcodes = await detector.detect(video);
                if (barcodes.length > 0) {
                    const isbn = barcodes[0].rawValue;
                    stopScanner();

                    // Show in manual input too
                    document.getElementById('manualISBN').value = isbn;

                    // Trigger lookup
                    if (typeof lookupAndShow === 'function') {
                        await lookupAndShow(isbn);
                    }
                }
            } catch (e) {
                // Detection failed, keep trying
            }
        }, 500);
    } catch (e) {
        alert('Could not access camera: ' + e.message);
    }
}

function stopScanner() {
    const container = document.getElementById('scannerContainer');
    const startBtn = document.getElementById('startScan');
    const stopBtn = document.getElementById('stopScan');

    if (scannerInterval) {
        clearInterval(scannerInterval);
        scannerInterval = null;
    }

    if (scannerStream) {
        scannerStream.getTracks().forEach(function(track) { track.stop(); });
        scannerStream = null;
    }

    container.classList.remove('active');
    startBtn.style.display = 'inline-flex';
    stopBtn.style.display = 'none';
}
