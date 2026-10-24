use std::{
    fs::{self, File, OpenOptions},
    io::Write,
    path::{Path, PathBuf},
};
use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum IdentityError {
    #[error("owner credential already exists")]
    AlreadyProvisioned,
    #[error("owner credential I/O failed: {0}")]
    Io(#[from] std::io::Error),
}

/// Creates a random owner bearer secret without replacing an existing credential.
/// The caller should run this as the account that owns the credential source file.
pub fn provision_owner_token(path: impl AsRef<Path>) -> Result<String, IdentityError> {
    let path = path.as_ref();
    let parent = path
        .parent()
        .filter(|p| !p.as_os_str().is_empty())
        .unwrap_or(Path::new("."));
    let parent_existed = parent.exists();
    fs::create_dir_all(parent)?;
    #[cfg(unix)]
    if !parent_existed {
        use std::os::unix::fs::PermissionsExt;
        fs::set_permissions(parent, fs::Permissions::from_mode(0o700))?;
    }

    let token = Uuid::new_v4().to_string();
    let temp_path = temporary_path(parent);
    let create_temp = || {
        let mut options = OpenOptions::new();
        options.write(true).create_new(true);
        #[cfg(unix)]
        {
            use std::os::unix::fs::OpenOptionsExt;
            options.mode(0o600);
        }
        options.open(&temp_path)
    };
    let mut temporary = create_temp()?;
    if let Err(error) = write_and_sync(&mut temporary, &token) {
        drop(temporary);
        let _ = fs::remove_file(&temp_path);
        return Err(error.into());
    }
    drop(temporary);
    if let Err(error) = fs::hard_link(&temp_path, path) {
        let _ = fs::remove_file(&temp_path);
        if error.kind() == std::io::ErrorKind::AlreadyExists {
            return Err(IdentityError::AlreadyProvisioned);
        }
        return Err(error.into());
    }
    fs::remove_file(&temp_path)?;
    sync_parent(parent)?;
    Ok(token)
}

fn temporary_path(parent: &Path) -> PathBuf {
    parent.join(format!(".luna-owner-{}.tmp", Uuid::new_v4()))
}

fn write_and_sync(file: &mut File, token: &str) -> std::io::Result<()> {
    file.write_all(token.as_bytes())?;
    file.write_all(b"\n")?;
    file.sync_all()
}

#[cfg(unix)]
fn sync_parent(parent: &Path) -> std::io::Result<()> {
    File::open(parent)?.sync_all()
}

#[cfg(not(unix))]
fn sync_parent(_: &Path) -> std::io::Result<()> {
    Ok(())
}
