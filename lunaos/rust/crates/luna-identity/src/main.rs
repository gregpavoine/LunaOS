use luna_identity::provision_owner_token;
use std::{path::PathBuf, process::ExitCode};

fn main() -> ExitCode {
    let mut path = PathBuf::from("/etc/lunaos/owner-token");
    let mut reveal = false;
    let mut args = std::env::args_os().skip(1);
    while let Some(arg) = args.next() {
        match arg.to_str() {
            Some("--path") => match args.next() {
                Some(value) => path = PathBuf::from(value),
                None => {
                    eprintln!("--path requires a file path");
                    return ExitCode::from(2);
                }
            },
            Some("--reveal-token") => reveal = true,
            _ => {
                eprintln!("usage: luna-owner-setup [--path FILE] [--reveal-token]");
                return ExitCode::from(2);
            }
        }
    }
    match provision_owner_token(&path) {
        Ok(token) => {
            println!("Owner credential provisioned at {}", path.display());
            if reveal {
                println!("{token}");
            }
            ExitCode::SUCCESS
        }
        Err(error) => {
            eprintln!("Owner credential provisioning failed: {error}");
            ExitCode::FAILURE
        }
    }
}
