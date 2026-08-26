const https = require('https');

async function getOpenRouterModels() {
    return new Promise((resolve) => {
        const req = https.request('https://openrouter.ai/api/v1/models', {
            method: 'GET',
        }, res => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                if (res.statusCode === 200) resolve(JSON.parse(body));
                else resolve(null);
            });
        });
        req.end();
    });
}

async function getGroqModels() {
    return new Promise((resolve) => {
        const req = https.request('https://api.groq.com/openai/v1/models', {
            method: 'GET',
            headers: { 'Authorization': `Bearer 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` }
        }, res => {
            let body = '';
            res.on('data', chunk => body += chunk);
            res.on('end', () => {
                if (res.statusCode === 200) resolve(JSON.parse(body));
                else resolve(null);
            });
        });
        req.end();
    });
}

async function main() {
    const groq = await getGroqModels();
    if (groq) {
        console.log("Groq Models:");
        groq.data.forEach(m => console.log(m.id));
    }
    
    const or = await getOpenRouterModels();
    if (or) {
        console.log("\nOpenRouter Free Models:");
        const freeModels = or.data.filter(m => m.pricing && m.pricing.prompt === "0" && m.pricing.completion === "0");
        freeModels.forEach(m => console.log(m.id));
    }
}
main();
