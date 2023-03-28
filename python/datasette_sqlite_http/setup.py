from setuptools import setup

version = {}
with open("datasette_sqlite_http/version.py") as fp:
    exec(fp.read(), version)

VERSION = version['__version__']

setup(
    name="datasette-sqlite-http",
    description="",
    long_description="",
    long_description_content_type="text/markdown",
    author="Alex Garcia",
    url="https://github.com/asg017/sqlite-http",
    project_urls={
        "Issues": "https://github.com/asg017/sqlite-http/issues",
        "CI": "https://github.com/asg017/sqlite-http/actions",
        "Changelog": "https://github.com/asg017/sqlite-http/releases",
    },
    license="MIT License, Apache License, Version 2.0",
    version=VERSION,
    packages=["datasette_sqlite_http"],
    entry_points={"datasette": ["sqlite_http = datasette_sqlite_http"]},
    install_requires=["datasette", "sqlite-http"],
    extras_require={"test": ["pytest"]},
    python_requires=">=3.7",
)