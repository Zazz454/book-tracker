// Lightweight chart rendering using Canvas API.
// No external dependencies.

function getChartColors() {
    const style = getComputedStyle(document.documentElement);
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches ||
                   document.documentElement.getAttribute('data-theme') === 'dark';
    return {
        text: isDark ? '#a0a0b0' : '#6e6e73',
        grid: isDark ? '#2e2e4a' : '#e5e5ea',
        bar: isDark ? '#4da6ff' : '#0071e3',
        palette: [
            '#0071e3', '#34c759', '#ff9500', '#ff3b30', '#af52de',
            '#5ac8fa', '#ffcc00', '#ff2d55', '#5856d6', '#64d2ff'
        ]
    };
}

function drawBarChart(canvasId, labels, values, label) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const colors = getChartColors();

    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    const w = rect.width;
    const h = rect.height;
    const padding = { top: 10, right: 10, bottom: 40, left: 40 };
    const chartW = w - padding.left - padding.right;
    const chartH = h - padding.top - padding.bottom;

    const max = Math.max(...values, 1);

    ctx.clearRect(0, 0, w, h);

    // Grid lines
    ctx.strokeStyle = colors.grid;
    ctx.lineWidth = 0.5;
    const gridLines = 4;
    for (let i = 0; i <= gridLines; i++) {
        const y = padding.top + chartH - (chartH * i / gridLines);
        ctx.beginPath();
        ctx.moveTo(padding.left, y);
        ctx.lineTo(padding.left + chartW, y);
        ctx.stroke();

        ctx.fillStyle = colors.text;
        ctx.font = '10px -apple-system, sans-serif';
        ctx.textAlign = 'right';
        ctx.fillText(Math.round(max * i / gridLines), padding.left - 5, y + 3);
    }

    // Bars
    const barWidth = Math.min(chartW / labels.length * 0.7, 40);
    const gap = chartW / labels.length;

    values.forEach(function(val, i) {
        const barH = (val / max) * chartH;
        const x = padding.left + gap * i + (gap - barWidth) / 2;
        const y = padding.top + chartH - barH;

        ctx.fillStyle = colors.bar;
        ctx.beginPath();
        const r = 3;
        ctx.moveTo(x + r, y);
        ctx.lineTo(x + barWidth - r, y);
        ctx.quadraticCurveTo(x + barWidth, y, x + barWidth, y + r);
        ctx.lineTo(x + barWidth, y + barH);
        ctx.lineTo(x, y + barH);
        ctx.lineTo(x, y + r);
        ctx.quadraticCurveTo(x, y, x + r, y);
        ctx.fill();

        // Label
        ctx.fillStyle = colors.text;
        ctx.font = '9px -apple-system, sans-serif';
        ctx.textAlign = 'center';
        ctx.save();
        ctx.translate(x + barWidth / 2, padding.top + chartH + 12);
        ctx.rotate(-0.4);
        const shortLabel = labels[i].length > 7 ? labels[i].substring(0, 7) : labels[i];
        ctx.fillText(shortLabel, 0, 0);
        ctx.restore();

        // Value on top
        if (val > 0) {
            ctx.fillStyle = colors.text;
            ctx.font = 'bold 10px -apple-system, sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText(val, x + barWidth / 2, y - 4);
        }
    });
}

function drawPieChart(canvasId, labels, values) {
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const colors = getChartColors();

    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    const w = rect.width;
    const h = rect.height;
    const total = values.reduce(function(a, b) { return a + b; }, 0);
    if (total === 0) return;

    const cx = w * 0.35;
    const cy = h / 2;
    const radius = Math.min(cx - 10, cy - 10);

    let startAngle = -Math.PI / 2;

    values.forEach(function(val, i) {
        const sliceAngle = (val / total) * Math.PI * 2;
        ctx.beginPath();
        ctx.moveTo(cx, cy);
        ctx.arc(cx, cy, radius, startAngle, startAngle + sliceAngle);
        ctx.closePath();
        ctx.fillStyle = colors.palette[i % colors.palette.length];
        ctx.fill();

        // Slight gap between slices
        ctx.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--bg-card').trim() || '#fff';
        ctx.lineWidth = 2;
        ctx.stroke();

        startAngle += sliceAngle;
    });

    // Inner circle for donut
    ctx.beginPath();
    ctx.arc(cx, cy, radius * 0.5, 0, Math.PI * 2);
    ctx.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--bg-card').trim() || '#fff';
    ctx.fill();

    // Legend
    const legendX = w * 0.65;
    let legendY = 15;
    ctx.font = '11px -apple-system, sans-serif';

    labels.forEach(function(label, i) {
        const pct = Math.round((values[i] / total) * 100);
        ctx.fillStyle = colors.palette[i % colors.palette.length];
        ctx.fillRect(legendX, legendY - 8, 10, 10);

        ctx.fillStyle = colors.text;
        ctx.textAlign = 'left';
        const shortLabel = label.length > 15 ? label.substring(0, 15) + '...' : label;
        ctx.fillText(shortLabel + ' (' + pct + '%)', legendX + 14, legendY);
        legendY += 18;
    });
}
