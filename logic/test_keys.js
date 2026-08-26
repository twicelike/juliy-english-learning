const https = require('https');
const http = require('http');

async function testGroq(apiKey) {
    return new Promise((resolve) => {
        if (!apiKey) { resolve("No Key"); return; }
        const data = JSON.stringify({
            model: "llama3-8b-8192",
            messages: [{ role: "user", content: "Hi" }]
        });
        const req = https.request('https://api.groq.com/openai/v1/chat/completions', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${apiKey}`, 'Content-Type': 'application/json' }
        }, res => {
            if (res.statusCode === 200) resolve("🟢 Groq: OK (Has Quota)");
            else resolve(`🔴 Groq: Failed (Status ${res.statusCode})`);
        });
        req.on('error', e => resolve(`🔴 Groq: Error - ${e.message}`));
        req.write(data);
        req.end();
    });
}

async function testHuggingFace(token) {
    return new Promise((resolve) => {
        if (!token) { resolve("No Key"); return; }
        const data = JSON.stringify({ inputs: "Hi", parameters: { max_new_tokens: 10 } });
        const req = https.request('https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.2', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' }
        }, res => {
            if (res.statusCode === 200) resolve("🟢 HuggingFace: OK (Has Quota)");
            else resolve(`🔴 HuggingFace: Failed (Status ${res.statusCode})`);
        });
        req.on('error', e => resolve(`🔴 HuggingFace: Error - ${e.message}`));
        req.write(data);
        req.end();
    });
}

async function testOpenRouter(apiKey) {
    return new Promise((resolve) => {
        if (!apiKey) { resolve("No Key"); return; }
        const data = JSON.stringify({
            model: "google/gemini-2.5-flash", // test with paid model
            messages: [{ role: "user", content: "Hi" }],
            max_tokens: 10
        });
        const req = https.request('https://openrouter.ai/api/v1/chat/completions', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${apiKey}`, 'Content-Type': 'application/json' }
        }, res => {
            if (res.statusCode === 200) resolve("🟢 OpenRouter (Paid Model): OK (Has Quota)");
            else resolve(`🔴 OpenRouter (Paid Model): Failed (Status ${res.statusCode}) - Out of credits`);
        });
        req.on('error', e => resolve(`🔴 OpenRouter: Error - ${e.message}`));
        req.write(data);
        req.end();
    });
}

async function testOpenRouterFree(apiKey) {
    return new Promise((resolve) => {
        if (!apiKey) { resolve("No Key"); return; }
        const data = JSON.stringify({
            model: "google/gemini-2.0-flash-lite-preview-02-05:free",
            messages: [{ role: "user", content: "Hi" }],
            max_tokens: 10
        });
        const req = https.request('https://openrouter.ai/api/v1/chat/completions', {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${apiKey}`, 'Content-Type': 'application/json' }
        }, res => {
            if (res.statusCode === 200) resolve("🟢 OpenRouter (Free Model): OK (Working)");
            else resolve(`🔴 OpenRouter (Free Model): Failed (Status ${res.statusCode})`);
        });
        req.on('error', e => resolve(`🔴 OpenRouter Free: Error - ${e.message}`));
        req.write(data);
        req.end();
    });
}

async function testOllama() {
    return new Promise((resolve) => {
        const req = http.request('http://localhost:11434/api/tags', { method: 'GET' }, res => {
            if (res.statusCode === 200) resolve("🟢 Ollama: OK (Local Server Running)");
            else resolve(`🔴 Ollama: Failed (Status ${res.statusCode})`);
        });
        req.on('error', e => resolve(`🔴 Ollama: Not Running (Localhost)`));
        req.end();
    });
}

async function main() {
    const groqKey = "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";
    const orKey = "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";
    const hfKey = "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";
    
    console.log("Testing AI Clients...\n");
    const results = await Promise.all([
        testGroq(groqKey),
        testHuggingFace(hfKey),
        testOpenRouter(orKey),
        testOpenRouterFree(orKey),
        testOllama()
    ]);
    
    results.forEach(r => {
        if (r !== "No Key") console.log(r);
    });
}

main();
