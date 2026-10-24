use luna_agent_runner::{ActionRunner, LocalActionRunner};

#[test]
fn checksum_action_returns_verifiable_public_fixture_evidence() {
    let evidence = LocalActionRunner.execute("local.checksum").unwrap();

    assert_eq!(
        evidence.sha256,
        "07be6a1cc7b1b323fd3baca70abf34bfea13e09f9d28cf417d11aca77d4220ce"
    );
    assert_eq!(evidence.byte_count, 28);
    assert_eq!(evidence.provenance, "local_public_fixture_v1");
}

#[test]
fn runner_rejects_actions_outside_its_whitelist() {
    assert!(LocalActionRunner.execute("system.remove_packages").is_err());
}
