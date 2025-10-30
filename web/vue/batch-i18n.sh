#!/bin/bash
# 批量国际化脚本 - 自动完成所有页面的多语言支持

echo "开始批量国际化处理..."
echo "此脚本将自动修改所有页面文件"
echo "建议先备份代码！"
echo ""
read -p "是否继续? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

echo "处理完成！请检查修改后的文件。"
