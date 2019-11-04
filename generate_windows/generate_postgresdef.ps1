# Generates interface library file for postgres.exe. 
# MSVC compiled file postgres.lib can't be used by gcc (gcc imports nothing from postgres.exe).
# Use genereted libpostgresexe.lib when linking with postgres.exe. 
# You need dlltool.exe in your path. Dlltool.exe is a part of gcc for mindows (mingw)
# You need dumpbin.exe.

$YourPathToPostgres = 'C:\Program Files\PostgreSQL\10.9-5.1C\'
$YourPathToMSVCbin = 'C:\Program Files (x86)\Microsoft Visual Studio\2017\Community\VC\Tools\MSVC\14.14.26428\bin\'

# creats 'exports definition file .def'
'LIBRARY postgres.exe
EXPORTS' | Out-File .\postgresExeExports.def -encoding utf8

echo 'running dumpbin.exe /exports '$YourPathToPostgres'bin\postgres.exe'
# // runs msvc's dumpbin.exe /exports to get exports of postgres.exe
& $YourPathToMSVCbin'Hostx64\x64\dumpbin.exe' /exports $YourPathToPostgres'bin\postgres.exe' | Select-Object -Skip 19 | Select-Object -SkipLast 8 | foreach { echo $_.ToString().SubString(26)} | Out-File .\postgresExeExports.def -Append -encoding utf8

echo 'running dlltool.exe -d .\postgresExeExports.def -l libpostgresInterfaceLib.a -D postgres.exe' $YourPathToPostgres'bin\postgres.exe'
# // generates 'interface library' from .def file to use with gcc mingw on Windows to link to postgres.exe
dlltool.exe -d .\postgresExeExports.def -l libpostgresInterfaceLib.a -D postgres.exe $YourPathToPostgres'bin\postgres.exe'