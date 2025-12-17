from playwright.sync_api import sync_playwright, expect
import os

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page(viewport={"width": 1280, "height": 1024})

    # Mock Wails Runtime and Backend
    page.add_init_script("""
        window.listeners = {};
        window.runtime = {
            EventsOn: (name, callback) => {
                if (!window.listeners[name]) window.listeners[name] = [];
                window.listeners[name].push(callback);
                return () => {};
            },
            EventsEmit: () => {},
            WindowUnminimise: () => {},
            WindowShow: () => {},
            WindowSetAlwaysOnTop: () => {}
        };
        window.go = {
            main: {
                App: {
                    GetContextMenuStatus: () => Promise.resolve(true),
                    GetThumbnail: () => Promise.resolve(""),
                    ConvertFiles: () => Promise.resolve(),
                    CopyFileToClipboard: () => Promise.resolve(),
                    CancelJob: () => Promise.resolve(),
                    PauseQueue: () => Promise.resolve(),
                    ResumeQueue: () => Promise.resolve()
                }
            }
        };
    """)

    try:
        page.goto("http://localhost:5173")

        # Verify initial state (DropZone)
        expect(page.locator("text=Drag & Drop files here")).to_be_visible()

        # Inject a file
        page.evaluate("""
            if (window.listeners['files-received']) {
                window.listeners['files-received'].forEach(cb => cb(['/tmp/test_video.mov', '/tmp/test_image.heic']));
            }
        """)

        # Wait for list to render
        item = page.get_by_text("test_video.mov")
        expect(item).to_be_visible()

        page.screenshot(path="verification/verification.png", full_page=True)
        print("Verification successful!")

    except Exception as e:
        print(f"Verification failed: {e}")
        page.screenshot(path="verification/error.png")
        raise e
    finally:
        browser.close()

with sync_playwright() as playwright:
    run(playwright)
