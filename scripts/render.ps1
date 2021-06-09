$project = "bad_apple"

$start = 1256428800000
$interval = 300000

1..5301 | ForEach-Object {
    $fstart = $start + ($interval * ($_ - 1))
    $fend = $fstart + $interval - 44000
    Write-Host "Rendering frame $_ : $fstart - $fend"

    $img = "{0:d6}.png" -f $_
    curl -o "frames/img/$project/$img" "http://admin:admin@localhost:3000/render/d/pvr-dash-8/prometheus-video-renderer-8-0?orgId=1&from=$fstart&to=$fend&panelId=2&width=1280&height=1100&timeout=120"
}
