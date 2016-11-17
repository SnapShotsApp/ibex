require 'rake/clean'

CLEAN << 'bindata.go'
CLEAN << 'build'

SRC = FileList['bindata.go', '*.go', 'Godeps/*', 'vendor/**/*']

directory 'build'

file 'bindata.go' => FileList['resources/*'] do
  sh 'go-bindata resources/'
end

file 'ibex-osx' => SRC do
  sh({ 'GOOS' => 'darwin' }, 'go build -o build/ibex-osx -v .')
end

file 'ibex-linux' => SRC do
  sh({ 'GOOS' => 'linux' }, 'go build -o build/ibex-linux -v .')
end

task all: ['ibex-osx', 'ibex-linux']
task default: :all

task run: 'ibex-osx' do
  sh './build/ibex', 'resources/config.json'
end
