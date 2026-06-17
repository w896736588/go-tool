#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
显示指定文件在当前分支中的改动内容

用法: python show_file_diff.py <基分支> <文件路径>

  基分支: 分支名（对比代码模式）或 "_workspace_"（工作区变更模式，对比 HEAD）
  工作区模式下，对比当前工作区文件与 HEAD 版本的差异。

输出 JSON:
- 文本文件: {"diff": "...", "old_content": "...", "new_content": "..."}
- 二进制文件: {"is_binary": true, "file_type": ".exe", "old_size": 12345, "new_size": 23456}
- 图片文件: {"is_image": true, "image_type": "png", "old_image": "base64...", "new_image": "base64..."}
"""

import base64
import json
import os
import re
import subprocess
import sys

# Windows 中文环境下 stdout 默认编码为 GBK，导致 UTF-8 输出乱码。
sys.stdout.reconfigure(encoding='utf-8', errors='replace')
sys.stderr.reconfigure(encoding='utf-8', errors='replace')

# 二进制文件扩展名集合
BINARY_EXTENSIONS = {
    '.exe', '.dll', '.so', '.dylib', '.bin', '.dat', '.zip', '.tar', '.gz',
    '.7z', '.rar', '.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx',
    '.ttf', '.otf', '.woff', '.woff2', '.eot', '.mp3', '.mp4', '.avi', '.mov',
    '.mkv', '.webm', '.wav', '.flac', '.ogg', '.o', '.a', '.lib', '.class',
    '.jar', '.war', '.pyc', '.pyo', '.wasm', '.ico', '.cur', '.db', '.sqlite',
    '.sqlite3', '.node', '.lock', '.sum', '.whl', '.tgz', '.tar.gz', '.tar.bz2',
    '.tar.xz', '.bz2', '.xz', '.lz4', '.zst', '.iso', '.dmg', '.pkg', '.deb',
    '.rpm', '.apk', '.ipa', '.msi', '.patch',
}

# 图片文件扩展名集合（SVG 除外，它是文本格式；ico 归到二进制因为大多数是二进制）
IMAGE_EXTENSIONS = {
    '.png', '.jpg', '.jpeg', '.gif', '.webp', '.bmp', '.tiff', '.tif', '.svg',
}


def run_git(*args: str) -> str:
    result = subprocess.run(
        ["git"] + list(args), capture_output=True, text=True,
        encoding="utf-8", errors="replace",
    )
    if result.returncode != 0:
        msg = result.stderr.strip()
        print(f"ERROR: {msg}", file=sys.stderr)
        sys.exit(1)
    return result.stdout.strip()


def run_git_safe(*args: str) -> str:
    """运行 git 命令，出错时返回空字符串而不退出。"""
    result = subprocess.run(
        ["git"] + list(args), capture_output=True, text=True,
        encoding="utf-8", errors="replace",
    )
    if result.returncode != 0:
        return ""
    return result.stdout


def is_excluded_file(file_path: str) -> bool:
    normalized = file_path.replace("\\", "/")
    return bool(re.search(r"(^|/)dist/", normalized))


def get_file_ext(file_path: str) -> str:
    """获取文件扩展名（小写），支持复合扩展名如 .tar.gz"""
    name = os.path.basename(file_path).lower()
    # 复合扩展名检测
    for compound in ('.tar.gz', '.tar.bz2', '.tar.xz'):
        if name.endswith(compound):
            return compound
    _, ext = os.path.splitext(name)
    return ext


def is_binary_ext(ext: str) -> bool:
    """判断扩展名是否为二进制类型（不含图片）"""
    return ext in BINARY_EXTENSIONS


def is_image_ext(ext: str) -> bool:
    """判断扩展名是否为图片类型"""
    return ext in IMAGE_EXTENSIONS


def is_binary_by_content(file_path: str) -> bool:
    """通过读取文件前几 KB 判断是否为二进制文件（检测 null 字节）"""
    try:
        with open(file_path, 'rb') as f:
            chunk = f.read(8192)
        return b'\x00' in chunk
    except Exception:
        return False


def get_binary_diff(old_ref: str, file_path: str, normalized_path: str, ext: str) -> dict:
    """获取二进制文件的变更信息（仅大小变化）。old_ref 为 merge-base 或 HEAD。"""
    old_size = 0
    new_size = 0

    # 获取旧版本大小
    old_size_result = subprocess.run(
        ["git", "cat-file", "-s", f"{old_ref}:{normalized_path}"],
        capture_output=True, text=True, encoding="utf-8", errors="replace",
    )
    if old_size_result.returncode == 0:
        try:
            old_size = int(old_size_result.stdout.strip())
        except ValueError:
            old_size = 0

    # 获取当前文件大小
    abs_path = os.path.abspath(file_path)
    if os.path.isfile(abs_path):
        try:
            new_size = os.path.getsize(abs_path)
        except Exception:
            new_size = 0

    return {
        "is_binary": True,
        "file_type": ext,
        "old_size": old_size,
        "new_size": new_size,
    }


def get_image_diff(old_ref: str, file_path: str, normalized_path: str, ext: str) -> dict:
    """获取图片文件的变更信息（base64 编码的新旧图片）。old_ref 为 merge-base 或 HEAD。"""
    old_image_b64 = ""
    new_image_b64 = ""

    # 获取旧版本图片（二进制）
    old_result = subprocess.run(
        ["git", "show", f"{old_ref}:{normalized_path}"],
        capture_output=True,
    )
    if old_result.returncode == 0 and old_result.stdout:
        old_image_b64 = base64.b64encode(old_result.stdout).decode('ascii')

    # 获取当前图片
    abs_path = os.path.abspath(file_path)
    if os.path.isfile(abs_path):
        try:
            with open(abs_path, 'rb') as f:
                new_data = f.read()
            new_image_b64 = base64.b64encode(new_data).decode('ascii')
        except Exception:
            pass

    # 去掉开头的点号作为 image_type
    image_type = ext.lstrip('.')

    return {
        "is_image": True,
        "image_type": image_type,
        "old_image": old_image_b64,
        "new_image": new_image_b64,
    }


def format_size(size_bytes: int) -> str:
    """格式化文件大小为人类可读格式"""
    if size_bytes < 1024:
        return f"{size_bytes} B"
    elif size_bytes < 1024 * 1024:
        return f"{size_bytes / 1024:.1f} KB"
    elif size_bytes < 1024 * 1024 * 1024:
        return f"{size_bytes / (1024 * 1024):.2f} MB"
    else:
        return f"{size_bytes / (1024 * 1024 * 1024):.2f} GB"


def main() -> int:
    if len(sys.argv) < 3:
        print("用法: python show_file_diff.py <基分支|_workspace_> <文件路径>", file=sys.stderr)
        sys.exit(1)

    base_branch = sys.argv[1].strip()
    file_path = sys.argv[2].strip()
    if not base_branch:
        print("基分支不能为空", file=sys.stderr)
        sys.exit(1)
    if not file_path:
        print("文件路径不能为空", file=sys.stderr)
        sys.exit(1)

    # 工作区模式：对比 HEAD 与当前工作区
    is_workspace_mode = (base_branch == "_workspace_")

    # 检查是否在 git 仓库中
    try:
        run_git("rev-parse", "--show-toplevel")
    except SystemExit:
        print("当前目录不是 git 仓库", file=sys.stderr)
        sys.exit(1)

    # 排除 dist 目录
    if is_excluded_file(file_path):
        print(f"文件 '{file_path}' 位于 dist 目录下，已按规则过滤", file=sys.stderr)
        sys.exit(1)

    normalized_path = file_path.replace("\\", "/")

    # 检测文件类型，决定处理方式
    ext = get_file_ext(file_path)

    if is_workspace_mode:
        # 工作区模式：使用 HEAD 作为旧版本基准
        old_ref = "HEAD"

        # 验证 HEAD 存在
        try:
            run_git("rev-parse", "--verify", "HEAD")
        except SystemExit:
            print("当前仓库没有任何提交（HEAD 不存在）", file=sys.stderr)
            sys.exit(1)
    else:
        # 对比代码模式：验证基分支存在
        try:
            run_git("rev-parse", "--verify", base_branch)
        except SystemExit:
            print(f"基分支 '{base_branch}' 不存在", file=sys.stderr)
            sys.exit(1)

        # 获取 merge-base
        old_ref = run_git("merge-base", base_branch, "HEAD")
        if not old_ref:
            print(f"无法计算 '{base_branch}' 与当前分支的 merge-base", file=sys.stderr)
            sys.exit(1)

    # 二进制文件：直接返回大小变化信息
    if is_binary_ext(ext):
        output = get_binary_diff(old_ref, file_path, normalized_path, ext)
        print(json.dumps(output, ensure_ascii=False))
        return 0

    # 图片文件：返回 base64 编码的图片数据
    if is_image_ext(ext):
        output = get_image_diff(old_ref, file_path, normalized_path, ext)
        print(json.dumps(output, ensure_ascii=False))
        return 0

    diff_content = ""

    if is_workspace_mode:
        # 工作区模式：合并暂存区 + 工作区变更
        parts = []
        # 1) 暂存区改动（已 git add 未 commit）
        result = subprocess.run(
            ["git", "diff", "--cached", "--", normalized_path],
            capture_output=True, text=True, encoding="utf-8", errors="replace",
        )
        if result.returncode == 0 and result.stdout and result.stdout.strip():
            parts.append(result.stdout)

        # 2) 工作区改动（未 git add）
        result = subprocess.run(
            ["git", "diff", "--", normalized_path],
            capture_output=True, text=True, encoding="utf-8", errors="replace",
        )
        if result.returncode == 0 and result.stdout and result.stdout.strip():
            parts.append(result.stdout)

        diff_content = "\n".join(parts)
    else:
        # 对比代码模式：合并已提交 + 暂存区 + 工作区变更
        parts = []
        # 1) 已提交的改动（merge_base vs HEAD）
        result = subprocess.run(
            ["git", "diff", old_ref, "HEAD", "--", normalized_path],
            capture_output=True, text=True, encoding="utf-8", errors="replace",
        )
        if result.returncode == 0 and result.stdout and result.stdout.strip():
            parts.append(result.stdout)

        # 2) 暂存区改动（已 git add 未 commit）
        result = subprocess.run(
            ["git", "diff", "--cached", "--", normalized_path],
            capture_output=True, text=True, encoding="utf-8", errors="replace",
        )
        if result.returncode == 0 and result.stdout and result.stdout.strip():
            parts.append(result.stdout)

        # 3) 工作区改动（未 git add）
        result = subprocess.run(
            ["git", "diff", "--", normalized_path],
            capture_output=True, text=True, encoding="utf-8", errors="replace",
        )
        if result.returncode == 0 and result.stdout and result.stdout.strip():
            parts.append(result.stdout)

        diff_content = "\n".join(parts)

    # 获取旧版本文件内容（old_content）
    old_content = run_git_safe("show", f"{old_ref}:{normalized_path}")

    # 获取当前工作区文件内容（new_content）
    new_content = ""
    abs_path = os.path.abspath(file_path)
    if os.path.isfile(abs_path):
        try:
            with open(abs_path, 'r', encoding='utf-8', errors='replace') as f:
                new_content = f.read()
        except Exception:
            new_content = ""

    output = {
        "diff": diff_content,
        "old_content": old_content,
        "new_content": new_content,
    }

    print(json.dumps(output, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
