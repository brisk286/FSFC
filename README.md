# 面向灾难的服务器文件存储

## GOAL：

- 轻量级 工具
- 自动（以指定区域内文件的变动，一变一存，做到文件都没保存，我能保存（面向缓存？））保存 
- 上传（在没有网络环境下先保存在本地，一有网络环境马上上传） 
- 下载（单个文件下载（同步，覆盖），全体文件恢复）
- 无需上手 只需设置完成即可运行
- 分版本保存
- 前端要求：保存日志 可视化文件图

情景：
word突然崩溃
打开后，无法恢复到从前
我想从云端恢复到10分钟前的版本