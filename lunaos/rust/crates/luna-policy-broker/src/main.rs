use anyhow::{Context, Result, bail};
use luna_policy_broker::{PolicyBroker, serve_socket};
use std::{
    os::unix::{ffi::OsStrExt, fs::MetadataExt},
    path::Path,
};

fn private_regular_file(path: &Path) -> Result<()> {
    let metadata =
        std::fs::symlink_metadata(path).context("cannot inspect owner credential file")?;
    if !metadata.file_type().is_file()
        || metadata.uid() != unsafe { libc::geteuid() }
        || metadata.mode() & 0o077 != 0
    {
        bail!("owner credential must be a regular owner-only file owned by this service user");
    }
    Ok(())
}

#[tokio::main]
async fn main() -> Result<()> {
    let credential =
        std::env::var_os("LUNA_OWNER_TOKEN_FILE").context("LUNA_OWNER_TOKEN_FILE is required")?;
    let socket =
        std::env::var_os("LUNA_POLICY_SOCKET").context("LUNA_POLICY_SOCKET is required")?;
    let credential = Path::new(&credential);
    let socket = Path::new(&socket);
    private_regular_file(credential)?;
    let token = std::fs::read_to_string(credential)?.trim().to_owned();
    if token.len() != 64 || !token.bytes().all(|byte| byte.is_ascii_hexdigit()) {
        bail!("owner credential must contain 32 random bytes encoded as 64 hexadecimal characters");
    }
    if socket.as_os_str().as_bytes().contains(&0) {
        bail!("policy socket path contains a NUL byte");
    }
    let parent = socket
        .parent()
        .context("policy socket must have a parent directory")?;
    let dir =
        std::fs::symlink_metadata(parent).context("policy socket parent must already exist")?;
    if !dir.file_type().is_dir()
        || dir.uid() != unsafe { libc::geteuid() }
        || dir.mode() & 0o077 != 0
    {
        bail!("policy socket parent must be an owner-only directory owned by this service user");
    }
    eprintln!("LunaOS Rust policy broker listening on its private Unix socket");
    serve_socket(socket, PolicyBroker::new(token, None)).await?;
    Ok(())
}
