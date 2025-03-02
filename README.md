# vs_export

A tool to generate `compile_commands.json` from Visual Studio solution files (VS2015/2017/2019/2022).

## Usage

```bash
vs_export [-s <path>] -c <configuration>

Options:
    -s  <path>          Path to .sln file (optional, auto-detect in build/ if not specified)
    -c  <configuration> Build configuration (e.g., "Debug|x64", default: "Debug|x64")
```

## Examples

### With CMake Project

If your project uses CMake, simply:
1. Generate VS solution
2. Run vs_export
```bash
cmake -S . -B build
vs_export
```

### Manual Usage

For non-CMake projects or when you need specific configuration:
```bash
vs_export -s MyProject.sln -c "Release|x64"
```

## Output

The tool generates a `compile_commands.json` file that is compatible with:
- clangd
- ccls
- Other LSP (Language Server Protocol) servers
- Various C++ analysis tools

## Alternative Method

If you're using CMake with Ninja generator, you can get `compile_commands.json` directly:
```bash
cmake -S . -B build -GNinja -DCMAKE_EXPORT_COMPILE_COMMANDS=1
cmake --build build
```

## Why vs_export?

- Simple to use
- Works with existing Visual Studio solutions
- No need to modify project files
- Supports modern Visual Studio versions
- Automatic solution file detection

```cmd
Usage: vs_export -s <path> -c <configuration>

Where:
            -s   path                        sln filename
            -c   configuration               project configuration,eg Debug|Win32.
                                             default Debug|Win32
```