from playwright.sync_api import Page, expect, sync_playwright

def verify_frontend_changes(page: Page):
    # 1. Arrange: Go to the verification page
    page.goto("http://localhost:5173/verification.html")

    # 2. Assert: Check DropZone padding (indirectly by height or visual)
    # We will rely on visual inspection via screenshot, but we can assert existence.
    dropzone = page.locator("text=Drag & Drop files here")
    expect(dropzone).to_be_visible()

    # 3. Assert: Check FileList separation
    queue_header = page.locator("h2", has_text="Queue")
    expect(queue_header).to_be_visible()

    completed_header = page.locator("h2", has_text="Completed")
    expect(completed_header).to_be_visible()

    # 4. Assert: Check "Waiting" status text
    # The first item has status 'queued', so it should display 'Waiting'
    waiting_badge = page.locator("text=Waiting").first
    expect(waiting_badge).to_be_visible()

    # 5. Assert: Check 'done' items are under Completed
    # We can check that the completed header is followed by done items
    # For now, just taking a screenshot is enough as the structure is visual.

    # 6. Screenshot
    page.screenshot(path="verification_frontend.png", full_page=True)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_frontend_changes(page)
            print("Verification script ran successfully.")
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="verification_error.png")
        finally:
            browser.close()
