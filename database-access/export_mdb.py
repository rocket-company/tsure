"""
Exporta todas as tabelas dos arquivos .mdb para CSV.

pandas_access depende de mdbtools (Linux-only). No Windows,
usa pyodbc com o driver Microsoft Access que ja esta instalado.

Uso:
    python export_mdb.py
"""

import os
import sys
from pathlib import Path

import pandas as pd
import pyodbc

BASE_DIR = Path(__file__).parent
OUTPUT_DIR = BASE_DIR / "exports"
OUTPUT_DIR.mkdir(exist_ok=True)

ACCESS_DRIVER = "Microsoft Access Driver (*.mdb, *.accdb)"


def list_tables(conn: pyodbc.Connection) -> list[str]:
    cursor = conn.cursor()
    return [
        row.table_name
        for row in cursor.tables(tableType="TABLE")
    ]


def read_table(conn: pyodbc.Connection, table: str) -> pd.DataFrame:
    cursor = conn.cursor()
    cursor.execute(f"SELECT * FROM [{table}]")
    columns = [col[0] for col in cursor.description]
    rows = cursor.fetchall()
    return pd.DataFrame.from_records(
        [list(row) for row in rows],
        columns=columns,
    )


def export_mdb(mdb_path: Path) -> None:
    print(f"\n{'='*60}")
    print(f"Arquivo : {mdb_path.name}")

    conn_str = (
        f"DRIVER={{{ACCESS_DRIVER}}};"
        f"DBQ={mdb_path};"
        "ExtendedAnsiSQL=1;"
    )
    conn = pyodbc.connect(conn_str, autocommit=True)

    tables = list_tables(conn)
    print(f"Tabelas : {len(tables)}")

    db_slug = mdb_path.stem.replace(" ", "_")
    errors = []

    for table in tables:
        try:
            df = read_table(conn, table)
            safe = table.replace(" ", "_").replace("/", "-")
            out = OUTPUT_DIR / f"{db_slug}__{safe}.csv"
            df.to_csv(out, index=False, encoding="utf-8-sig")
            print(f"  OK  {table:<40} {len(df):>6} linhas  ->  {out.name}")
        except Exception as exc:
            print(f"  ERR {table:<40} {exc}")
            errors.append((table, exc))

    conn.close()

    if errors:
        print(f"\n{len(errors)} tabela(s) com erro.")
    else:
        print(f"\nExportacao concluida -> {OUTPUT_DIR}")


def main() -> None:
    mdb_files = sorted(BASE_DIR.glob("*.mdb"))

    if not mdb_files:
        print("Nenhum arquivo .mdb encontrado em", BASE_DIR)
        sys.exit(1)

    for mdb_path in mdb_files:
        export_mdb(mdb_path)


if __name__ == "__main__":
    main()
