use luna_identity::provision_owner_token;
use std::{fs, path::PathBuf};
use uuid::Uuid;

fn credential_path() -> PathBuf {
    std::env::temp_dir().join(format!("luna-owner-{}.token", Uuid::new_v4()))
}

#[test]
fn owner_token_is_created_once_with_private_permissions() {
    let path = credential_path();
    let token = provision_owner_token(&path).unwrap();

    assert_eq!(fs::read_to_string(&path).unwrap().trim(), token);
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        assert_eq!(
            fs::metadata(&path).unwrap().permissions().mode() & 0o777,
            0o600
        );
    }
    assert!(provision_owner_token(&path).is_err());
    let _ = fs::remove_file(path);
}
