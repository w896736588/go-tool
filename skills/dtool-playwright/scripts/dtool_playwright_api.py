import json
from urllib import error, request


def _post_json(base_url: str, token: str, path: str, payload: dict):
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = request.Request(
        url=f"{base_url}{path}",
        data=body,
        headers={
            "Content-Type": "application/json; charset=utf-8",
            "Token": token,
        },
        method="POST",
    )
    with request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode("utf-8"))


def browser_profile_open(base_url: str, token: str, smart_link_id: int, label: str,
                         account: dict | None = None, open_type: int = 0,
                         reuse_if_open: bool = True):
    payload = {
        "smart_link_id": smart_link_id,
        "label": label,
        "account": account or {},
        "open_type": open_type,
        "reuse_if_open": reuse_if_open,
    }
    return _post_json(base_url, token, "/api/ai/browser/session/open", payload)


def extract_user_data_dir(open_result: dict) -> str:
    return open_result.get("data", {}).get("user_data_dir", "")


def build_python_playwright_snippet(user_data_dir: str, goto_url: str = "") -> str:
    goto_line = ""
    if goto_url:
        goto_line = f'    page.goto(r"{goto_url}")\n'
    return (
        "from playwright.sync_api import sync_playwright\n\n"
        f'user_data_dir = r"{user_data_dir}"\n\n'
        "with sync_playwright() as p:\n"
        "    context = p.chromium.launch_persistent_context(\n"
        "        user_data_dir=user_data_dir,\n"
        "        headless=False,\n"
        "        no_viewport=True,\n"
        "    )\n"
        "    page = context.pages[0] if context.pages else context.new_page()\n"
        f"{goto_line}"
        "    print(page.title())\n"
    )


if __name__ == "__main__":
    base_url = "http://127.0.0.1:17170"
    token = "请替换为真实Token"

    try:
        open_result = browser_profile_open(
            base_url=base_url,
            token=token,
            smart_link_id=12,
            label="登录后首页",
            account={"user_name": "tester"},
        )
        print("open_result:")
        print(json.dumps(open_result, ensure_ascii=False, indent=2))

        user_data_dir = extract_user_data_dir(open_result)
        if user_data_dir:
            goto_url = open_result.get("data", {}).get("site", {}).get("url", "")
            print("\npython_playwright_snippet:")
            print(build_python_playwright_snippet(user_data_dir, goto_url))

    except error.HTTPError as exc:
        print(f"HTTP {exc.code} 失败: {exc.read().decode('utf-8', errors='replace')}")
    except Exception as exc:
        print(f"请求失败: {exc}")
