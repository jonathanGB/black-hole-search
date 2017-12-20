require 'csv'

results = []
print "Start ring size: " ; ringSize = gets.chomp.to_i
print "Increment: "; increment = gets.chomp.to_i
print "Max ring size: " ; maxRingSize = gets.chomp.to_i

%x(gsed -i 's/1000,/#{ringSize},/g' main_test.go) # replace ringSize in benchmark file

while ringSize <= maxRingSize
    puts ringSize # to visualize progress
    matches = %x(go test -bench=.).scan(/Benchmark\w+\d+\-\d\s+\d+\s+(\d+)\sns\/op\n/)
    results << matches.flatten.map(&:to_i).insert(0, ringSize)

    newRingSize = ringSize + increment
    %x(gsed -i 's/#{ringSize},/#{newRingSize},/g' main_test.go) # replace ringSize in benchmark file
    ringSize = newRingSize
end

%x(gsed -i 's/#{ringSize},/1000,/g' main_test.go) # replace ringSize in benchmark file

# write results to CSV file
results.insert(0, ["Ring Size", 'Divide', 'Group', 'OptAvgTime', 'OptTeamSize', 'OptTime'])
File.write("results.csv", results.map(&:to_csv).join)
