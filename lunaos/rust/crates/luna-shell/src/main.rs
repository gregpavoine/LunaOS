use gtk::{Application, ApplicationWindow, Label, glib, prelude::*};

fn main() -> glib::ExitCode {
    let app = Application::builder()
        .application_id("org.lunaos.Shell")
        .build();
    app.connect_activate(|app| {
        let status = Label::new(Some("Connecting to LUNA Core…"));
        status.set_margin_top(32); status.set_margin_bottom(32); status.set_margin_start(32); status.set_margin_end(32);
        let window = ApplicationWindow::builder().application(app).title("LunaOS").default_width(720).default_height(480).child(&status).build();
        window.present();
        let status_ref = status.clone();
        glib::MainContext::default().spawn_local(async move {
            let text = match reqwest::get("http://127.0.0.1:7766/v1/system/capabilities").await {
                Ok(response) => match response.error_for_status() { Ok(response) => response.text().await.unwrap_or_else(|_| "Core returned an unreadable response".into()), Err(_) => "LUNA Core returned an error".into() },
                Err(_) => "LUNA Core is unavailable. Start the luna-core systemd service to view live capability status.".into(),
            };
            status_ref.set_text(&text);
        });
    });
    app.run()
}
