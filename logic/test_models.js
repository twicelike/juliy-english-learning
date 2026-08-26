const https = require('https');

async function testOpenRouterFree(modelName) {
    return new Promise((resolve) => {
        const data = JSON.stringify({
            model: modelName,
            messages: [{ role: "user", content: "Hi" }],
            max_tokens: 2000
        });
        const req = https.request('https://openrouter.ai/api/v1/chat/completions', {
            method: 'POST',
            headers: { 
                'Authorization': `Bearer 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`, 
                'Content-Type': 'application/json' 
            }
        }, res => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                if (res.statusCode === 200) resolve(`🟢 ${modelName}: OK`);
                else resolve(`🔴 ${modelName}: Failed (${res.statusCode}) - ${body.substring(0, 150)}`);
            });
        });
        req.on('error', e => resolve(`🔴 ${modelName}: Error - ${e.message}`));
        req.write(data);
        req.end();
    });
}

async function testGroq() {
    return new Promise((resolve) => {
        const data = JSON.stringify({
            model: "llama3-8b-8192",
            messages: [{ role: "user", content: "Hi" }]
        });
        const req = https.request('https://api.groq.com/openai/v1/chat/completions', {
            method: 'POST',
            headers: { 
                'Authorization': `Bearer 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`, 
                'Content-Type': 'application/json' 
            }
        }, res => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                if (res.statusCode === 200) resolve(`🟢 Groq: OK`);
                else resolve(`🔴 Groq: Failed (${res.statusCode}) - ${body.substring(0, 150)}`);
            });
        });
        req.on('error', e => resolve(`🔴 Groq: Error - ${e.message}`));
        req.write(data);
        req.end();
    });
}

async function main() {
    const models = [
        "google/gemini-2.0-flash-exp:free",
        "google/gemini-2.0-flash-lite-preview-02-05:free",
        "google/gemini-2.5-flash:free",
        "meta-llama/llama-3-8b-instruct:free"
    ];
    for (const m of models) {
        console.log(await testOpenRouterFree(m));
    }
    console.log(await testGroq());
}
main();
