require 'rake/clean'

CLEAN << 'bindata.go'
CLEAN << 'ibex'

file 'bindata.go' => FileList['resources/*'] do
  sh 'go-bindata resources/'
end

file 'ibex' => FileList['bindata.go', '*.go', 'Godeps/*', 'vendor/**/*'] do
  sh 'go build .'
end

task default: 'ibex'

task run: 'ibex' do
  sh './ibex', 'resources/config.json'
end
