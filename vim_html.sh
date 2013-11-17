export output="`pwd`/html/code/"
export src="`pwd`/src/"
export home=`pwd`

cd src
export projects=`find . -mindepth 1 -type d -not -path '*/.git*'`

cd $output
for p in $projects;
do
	mkdir -p $p
done

cd $src
for p in $projects;
do
	for f in `find $p -maxdepth 1 -type f -printf "%p\n"`;
	do
		file=${f#./}
		out=$output$file

		echo $file
		#echo $file
		gvim -u $home/htmlrc.vim -E -s -c +":syn on" +"run! syntax/2html.vim" +"saveas! $out.html" +"qall!" $file >> /dev/null &
	done
	echo
done

