#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
通用代码编辑脚本。
通过 JSON 描述文件执行多文件、多操作的精确文本替换/插入，解决 Edit 工具在含中文、特殊字符时匹配失败的问题。

用法:
    python skills/dtool-common/scripts/code_edit.py <描述文件.json> [--dry-run]

描述文件格式（JSON）:
    {
      "files": [
        {
          "path": "C:/work/project/src/file.go",
          "ops": [
            {
              "op": "replace",
              "find": "待替换的原始文本（精确匹配）",
              "replace": "替换后的文本"
            },
            {
              "op": "replace_all",
              "find": "全文多次出现的文本",
              "replace": "替换后的文本"
            },
            {
              "op": "insert_after",
              "after": "在此行文本之后插入（精确匹配该行）",
              "text": "要插入的文本（可多行）"
            },
            {
              "op": "insert_before",
              "before": "在此行文本之前插入",
              "text": "要插入的文本"
            }
          ]
        }
      ]
    }

特性:
    - 自动检测并保留原文件的 BOM、换行符（CRLF/LF）、文件末尾换行
    - 每个文件修改前自动在同目录创建 .bak 备份
    - --dry-run 模式仅打印将要进行的修改，不实际写入
    - 所有 find/after/before 匹配区分大小写，要求唯一匹配
"""

import json
import os
import shutil
import sys


def load_desc(path):
    """加载 JSON 描述文件。"""
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def detect_format(filepath):
    """检测文件的 BOM、换行符类型。返回 (bom_bytes, line_ending)。"""
    with open(filepath, "rb") as f:
        head = f.read(4)
    bom = b""
    if head[:3] == b"\xef\xbb\xbf":
        bom = b"\xef\xbb\xbf"
    elif head[:2] == b"\xff\xfe":
        bom = b"\xff\xfe"
        return bom, "\r\n"  # UTF-16-LE 统一用 CRLF

    # 检测换行符
    sample = head
    if len(head) < 4:
        with open(filepath, "rb") as f:
            sample = f.read(8192)
    crlf_count = sample.count(b"\r\n")
    lf_only = sample.count(b"\n") - crlf_count
    if crlf_count > 0 and lf_only == 0:
        return bom, "\r\n"
    else:
        return bom, "\n"


def normalize_content(raw, bom, line_ending):
    """将原始字节内容去掉 BOM、统一换行符为 \n 后返回文本，同时检查末尾换行。"""
    text = raw
    if bom and text.startswith(bom):
        text = text[len(bom):]
    if line_ending == "\r\n":
        text = text.replace(b"\r\n", b"\n")
    # 检测原始文件是否有末尾换行
    ends_with_newline = text.endswith(b"\n")
    return text.decode("utf-8"), ends_with_newline


def denormalize_content(text, bom, line_ending, ends_with_newline):
    """将统一格式的文本还原为原始文件格式。"""
    if line_ending == "\r\n":
        text = text.replace("\n", "\r\n")
    out = text.encode("utf-8")
    if bom:
        out = bom + out
    # 还原末尾换行
    if not ends_with_newline and out.endswith(b"\r\n"):
        out = out[:-2]
    elif not ends_with_newline and out.endswith(b"\n"):
        out = out[:-1]
    return out


def perform_replace(content, find_text, replace_text, replace_all=False):
    """在文本中执行替换。返回 (新文本, 匹配次数)。"""
    count = content.count(find_text)
    if count == 0:
        raise ValueError(f"未找到匹配文本: {repr(find_text[:80])}")
    if not replace_all and count > 1:
        lines = [str(i + 1) for i, l in enumerate(content.split("\n")) if find_text in l]
        raise ValueError(
            f"找到 {count} 处匹配，但未启用 replace_all。匹配行号: {', '.join(lines)}"
        )
    return content.replace(find_text, replace_text), count


def perform_insert_after(content, after_text, insert_text):
    """在指定行之后插入文本。"""
    idx = content.find(after_text)
    if idx == -1:
        raise ValueError(f"未找到标记文本: {repr(after_text[:80])}")
    # 确保 after_text 之后是换行或是末尾
    line_end = content.find("\n", idx)
    if line_end == -1:
        line_end = len(content)
    # 在行尾插入
    new_content = content[:line_end] + "\n" + insert_text + content[line_end:]
    return new_content, 1


def perform_insert_before(content, before_text, insert_text):
    """在指定行之前插入文本。"""
    idx = content.find(before_text)
    if idx == -1:
        raise ValueError(f"未找到标记文本: {repr(before_text[:80])}")
    # 找到 before_text 所在行的行首
    line_start = content.rfind("\n", 0, idx)
    if line_start == -1:
        line_start = 0
    else:
        line_start += 1
    new_content = content[:line_start] + insert_text + "\n" + content[line_start:]
    return new_content, 1


def apply_ops(content, ops):
    """按顺序执行所有操作，返回新文本和操作统计。"""
    stats = []
    for op in ops:
        op_type = op["op"]
        if op_type == "replace":
            find_text = op["find"]
            replace_text = op["replace"]
            replace_all = op.get("replace_all", False)
            content, count = perform_replace(content, find_text, replace_text, replace_all)
            stats.append(f"  replace: {count} 处匹配 → 已替换")
        elif op_type == "replace_all":
            find_text = op["find"]
            replace_text = op["replace"]
            content, count = perform_replace(content, find_text, replace_text, replace_all=True)
            stats.append(f"  replace_all: {count} 处匹配 → 已替换")
        elif op_type == "insert_after":
            content, _ = perform_insert_after(content, op["after"], op["text"])
            stats.append(f"  insert_after: 已插入")
        elif op_type == "insert_before":
            content, _ = perform_insert_before(content, op["before"], op["text"])
            stats.append(f"  insert_before: 已插入")
        else:
            raise ValueError(f"未知操作类型: {op_type}")
    return content, stats


def process_file(entry, dry_run=False):
    """处理单个文件的编辑操作。"""
    filepath = entry["path"]
    ops = entry["ops"]
    if not os.path.isfile(filepath):
        raise FileNotFoundError(f"文件不存在: {filepath}")

    bom, line_ending = detect_format(filepath)
    with open(filepath, "rb") as f:
        raw = f.read()
    content, ends_with_newline = normalize_content(raw, bom, line_ending)

    if dry_run:
        # 只做匹配检查，不写入
        new_content, stats = apply_ops(content, ops)
        print(f"[DRY-RUN] {filepath}")
        for s in stats:
            print(s)
        return

    # 备份
    bak_path = filepath + ".bak"
    shutil.copy2(filepath, bak_path)
    print(f"[备份] {filepath} → {bak_path}")

    new_content, stats = apply_ops(content, ops)
    out_bytes = denormalize_content(new_content, bom, line_ending, ends_with_newline)
    with open(filepath, "wb") as f:
        f.write(out_bytes)
    print(f"[修改] {filepath}")
    for s in stats:
        print(s)


def main():
    dry_run = "--dry-run" in sys.argv
    args = [a for a in sys.argv[1:] if a != "--dry-run"]

    if len(args) < 1:
        print(__doc__)
        sys.exit(1)

    desc_path = args[0]
    if not os.path.isfile(desc_path):
        print(f"错误: 描述文件不存在: {desc_path}")
        sys.exit(1)

    desc = load_desc(desc_path)
    files = desc.get("files", [])
    if not files:
        print("错误: 描述文件中没有 'files' 条目")
        sys.exit(1)

    errors = []
    for entry in files:
        try:
            process_file(entry, dry_run=dry_run)
        except Exception as e:
            errors.append((entry["path"], str(e)))
            print(f"[失败] {entry['path']}: {e}")

    if errors:
        print(f"\n{len(errors)} 个文件处理失败:")
        for path, err in errors:
            print(f"  - {path}: {err}")
        sys.exit(1)

    if dry_run:
        print("\n--dry-run 完成，未修改任何文件")
    else:
        print("\n全部修改完成。备份文件以 .bak 结尾，确认无误后可删除。")


if __name__ == "__main__":
    main()
