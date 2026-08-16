# 置于 agent/ 根目录：pytest 以 prepend 模式将 conftest 所在目录加入 sys.path，
# 使 tests/ 内可直接 import guard、agent、memory、ws、config 等顶层包。
