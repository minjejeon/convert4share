
from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch()
    page = browser.new_page()
    page.on('console', lambda msg: print(f'CONSOLE: {msg.text}'))
    page.on('pageerror', lambda exc: print(f'PAGE ERROR: {exc}'))
    try:
        page.goto('http://localhost:5173')
        # Wait for something that is definitely there
        page.wait_for_selector('div', timeout=5000)
    except Exception as e:
        print(f'Test failed: {e}')

    page.screenshot(path='verification/frontend.png')
    browser.close()

with sync_playwright() as playwright:
    run(playwright)
