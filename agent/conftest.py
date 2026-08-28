# 置于 agent/ 根目录：显式将本目录加入 sys.path，
# 使 tests/ 内可直接 import guard、agent、memory、ws、config 等顶层包。
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
