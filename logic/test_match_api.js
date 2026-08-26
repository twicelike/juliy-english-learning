const http = require('http');

const req = http.request({
    hostname: 'localhost',
    port: 8080,
    path: '/api/v1/match/recommendations?limit=10',
    method: 'GET',
    headers: {
        // Need a valid auth token? The endpoint requires authentication.
        // Let's check if the endpoint uses a hardcoded test token or we need a real one.
    }
}, res => {
    let body = '';
    res.on('data', d => body += d);
    res.on('end', () => {
        console.log("Status:", res.statusCode);
        console.log("Body:", body);
    });
});
req.on('error', console.error);
req.end();
