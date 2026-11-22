param([Parameter(Mandatory=$true)][string]$OutputDirectory)
$ErrorActionPreference='Stop'
Add-Type -AssemblyName System.Data
if (-not $env:GOOSHISHOP_SQL_CONNECTION) {throw 'Set GOOSHISHOP_SQL_CONNECTION privately in this session.'}
New-Item -ItemType Directory -Force -Path $OutputDirectory | Out-Null
$conn=New-Object System.Data.SqlClient.SqlConnection
$builder=New-Object System.Data.SqlClient.SqlConnectionStringBuilder($env:GOOSHISHOP_SQL_CONNECTION)
$builder.PersistSecurityInfo=$false
$builder.MaxPoolSize=2
$builder.ApplicationName='CherrivaCatalogReadOnlyExport'
$conn.ConnectionString=$builder.ConnectionString
try {
 $conn.Open()
 foreach($file in Get-ChildItem (Join-Path $PSScriptRoot 'queries') -Filter '*.sql' | Sort-Object Name) {
  $cmd=$conn.CreateCommand();$cmd.CommandTimeout=120
  $cmd.CommandText=[IO.File]::ReadAllText($file.FullName)
  $reader=$cmd.ExecuteReader()
  $rows=New-Object 'System.Collections.Generic.List[object]'
  while($reader.Read()) {
   $row=[ordered]@{}
   for($i=0;$i -lt $reader.FieldCount;$i++) {
    $v=$reader.GetValue($i)
    if($v -is [DBNull]) {$v=$null}
    elseif($v -is [datetime]) {$v=$v.ToString('o')}
    $row[$reader.GetName($i)]=$v
   }
   $rows.Add($row)
  }
  $reader.Close();$cmd.Dispose()
  $json=ConvertTo-Json -InputObject @($rows.ToArray()) -Depth 30 -Compress
  [IO.File]::WriteAllText((Join-Path $OutputDirectory ($file.BaseName+'.json')),$json,(New-Object Text.UTF8Encoding($false)))
  Write-Output ($file.BaseName+': '+$rows.Count+' rows')
 }
} catch {throw 'Catalog extraction failed. Connection details were not logged.'}
finally {$conn.Dispose();$builder.Clear();Remove-Item Env:GOOSHISHOP_SQL_CONNECTION -ErrorAction SilentlyContinue}
