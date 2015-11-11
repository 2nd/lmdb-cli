require 'oj'
require 'lmdb'
require 'fileutils'

Oj.default_options = {mode: :compat}

template = './test/template/'
FileUtils.rm_r(template) if File.exists?(template)
FileUtils.mkdir(template)

env = LMDB.new(template)
db0 = env.database
db0['over'] = "9000!!"
24.times do |i|
  db0["iter:#{i}"] = "value-#{i}"
end
env.close()
