use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum LlmError {
    #[error("local llama.cpp backend is unavailable: {0}")]
    Unavailable(String),
    #[error("local model returned an invalid response")]
    InvalidResponse,
    #[error("llama.cpp endpoint must be an HTTP loopback URL")]
    NonLocalEndpoint,
}

#[derive(Clone)]
pub struct LlamaCppBackend {
    endpoint: String,
    model: String,
    client: reqwest::Client,
}

#[derive(Serialize)]
struct ChatRequest<'a> {
    model: &'a str,
    messages: [ChatMessage<'a>; 1],
    stream: bool,
}
#[derive(Serialize)]
struct ChatMessage<'a> {
    role: &'a str,
    content: &'a str,
}
#[derive(Deserialize)]
struct ChatResponse {
    choices: Vec<Choice>,
}
#[derive(Deserialize)]
struct Choice {
    message: Message,
}
#[derive(Deserialize)]
struct Message {
    content: String,
}

impl LlamaCppBackend {
    /// Calls the real OpenAI-compatible llama.cpp server. This is only inference
    /// when a local server is configured with an actual model file.
    pub fn new(endpoint: impl Into<String>, model: impl Into<String>) -> Result<Self, LlmError> {
        let endpoint = endpoint.into();
        let url = reqwest::Url::parse(&endpoint).map_err(|_| LlmError::NonLocalEndpoint)?;
        let is_loopback = matches!(url.host_str(), Some("localhost" | "127.0.0.1" | "::1"));
        if url.scheme() != "http"
            || !is_loopback
            || !url.username().is_empty()
            || url.password().is_some()
        {
            return Err(LlmError::NonLocalEndpoint);
        }
        Ok(Self {
            endpoint,
            model: model.into(),
            client: reqwest::Client::new(),
        })
    }

    pub async fn generate(&self, prompt: &str) -> Result<String, LlmError> {
        let url = format!(
            "{}/v1/chat/completions",
            self.endpoint.trim_end_matches('/')
        );
        let response = self
            .client
            .post(url)
            .json(&ChatRequest {
                model: &self.model,
                messages: [ChatMessage {
                    role: "user",
                    content: prompt,
                }],
                stream: false,
            })
            .send()
            .await
            .map_err(|e| LlmError::Unavailable(e.to_string()))?;
        if !response.status().is_success() {
            return Err(LlmError::Unavailable(format!(
                "llama.cpp HTTP {}",
                response.status()
            )));
        }
        let response: ChatResponse = response
            .json()
            .await
            .map_err(|_| LlmError::InvalidResponse)?;
        response
            .choices
            .into_iter()
            .next()
            .map(|choice| choice.message.content)
            .ok_or(LlmError::InvalidResponse)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn adapter_rejects_external_endpoints() {
        assert!(matches!(
            LlamaCppBackend::new("https://example.com", "model"),
            Err(LlmError::NonLocalEndpoint)
        ));
    }
}
