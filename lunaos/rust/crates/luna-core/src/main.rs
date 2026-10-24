use anyhow::{Context, Result};
use luna_core::{open_engine, router};
use tokio::net::TcpListener;

#[tokio::main]
async fn main() -> Result<()> {
    let token = if let Ok(path) = std::env::var("LUNA_OWNER_TOKEN_FILE") {
        std::fs::read_to_string(path)
            .context("could not read systemd-provided owner credential")?
            .trim()
            .to_owned()
    } else {
        std::env::var("LUNA_OWNER_TOKEN")
            .context("LUNA_OWNER_TOKEN_FILE must be supplied by the local credential store")?
    };
    let db = std::env::var("LUNA_DATABASE").unwrap_or_else(|_| "./lunaos-state.sqlite".into());
    let addr = std::env::var("LUNA_LISTEN").unwrap_or_else(|_| "127.0.0.1:7766".into());
    let engine = open_engine(&db, &token)?;
    let listener = TcpListener::bind(&addr).await?;
    eprintln!("LUNA Core listening on {addr}; database is local SQLite");
    let background_engine = engine.clone();
    tokio::spawn(async move {
        let mut scheduler = tokio::time::interval(std::time::Duration::from_secs(2));
        loop {
            scheduler.tick().await;
            if let Err(_error) = background_engine.tick() {
                eprintln!(
                    "LUNA scheduler deferred queued work after an authorization or storage failure"
                );
            }
        }
    });
    axum::serve(listener, router(engine, token)).await?;
    Ok(())
}
