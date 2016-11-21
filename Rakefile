require 'rake/clean'

CLEAN << 'bindata.go'
CLEAN << 'build'

SRC = FileList['bindata.go', '*.go', 'Godeps/*', 'vendor/**/*']

LICENSE_HEADER = %{
/* Copyright <%= @year %> Snapshots LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
}.strip.freeze

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

desc 'Build binaries for OS X and Linux'
task all: ['ibex-osx', 'ibex-linux']
task default: :all

task run: 'ibex-osx' do
  cmd = ['./build/ibex-osx']
  cmd << '--debug' if ENV['DEBUG']

  sh(*cmd, 'resources/config.json')
end

desc 'Runs all tests'
task test: 'bindata.go' do
  sh 'go', 'test'
end

desc 'Prepends the Apache License header comment to all sources'
task :license_all_files do
  require 'erb'
  require 'date'

  @year = Date.today.year
  license_text = ERB.new(LICENSE_HEADER).result binding

  sources = FileList.new('*.go') { |fl| fl.exclude 'bindata.go' }
  sources.each do |file|
    txt = File.read(file)
    next if txt =~ /Copyright.*Snapshots/
    File.write(file, [license_text, txt].join("\n\n"))
  end
end
