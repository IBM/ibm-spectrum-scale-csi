#!/bin/bash

TOTALCOUNT_POS=11
ERRORCOUNT_POS=13
AVGTIME_POS=15
MINTIME_POS=17
MAXTIME_POS=19

csiOpStats=(
"CreateVolume" 
"DeleteVolume" 
"CreateSnapshot" 
"DeleteSnapshot"
)

restAPIStats=(
GetClusterId
GetClusterSummary
GetTimeZoneOffset
GetScaleVersion
GetFilesystemMountDetails
IsFilesystemMountedOnGUINode
ListFilesystems
GetFilesystemDetails
GetFilesystemMountpoint
CreateFileset
UpdateFileset
DeleteFileset
LinkFileset
UnlinkFileset
ListFilesetDetails
GetFilesetsInodeSpace
IsFilesetLinked
FilesetRefreshTask
ListFilesetQuota
GetFilesetQuotaDetails
SetFilesetQuota
CheckIfFSQuotaEnabled
CheckIfFilesetExist
MakeDirectory
MakeDirectoryV2
MountFilesystem
UnmountFilesystem
GetFilesystemName
CheckIfFileDirPresent
CreateSymLink
GetFsUid
DeleteDirectory
StatDirectory
GetFileSetUid
GetFileSetNameFromId
DeleteSymLnk
GetFileSetResponseFromId
GetFileSetResponseFromName
MountFilesystem
UnmountFilesystem
GetFilesystemName
CheckIfFileDirPresent
CreateSymLink
GetFsUid
DeleteDirectory
StatDirectory
GetFileSetUid
GetFileSetNameFromId
DeleteSymLnk
GetFileSetResponseFromId
GetFileSetResponseFromName
SetFilesystemPolicy
DoesTierExist
GetTierInfoFromName
GetFirstDataTier
IsValidNodeclass
IsSnapshotSupported
CheckIfDefaultPolicyPartitionExists
WaitForJobCompletion
WaitForJobCompletionWithResp
CreateSnapshot
DeleteSnapshot
GetLatestFilesetSnapshots
GetSnapshotUid
GetSnapshotCreateTimestamp
CheckIfSnapshotExist
ListFilesetSnapshots
CopyFsetSnapshotPath
CopyFilesetPath
CopyDirectoryPath
IsNodeComponentHealthy
)

while getopts d: flag
do
    case "${flag}" in
        d) duration=${OPTARG};;
    esac
done

diffTime=$(date --date "-${duration} min" -u +%H%M%S)

podArr=( $(oc get pods -n ibm-spectrum-scale-csi | grep ibm-spectrum-scale-csi- | grep -v "operator\|attacher\|provisioner\|resizer\|snapshotter" | cut -d ' ' -f 1) )

mkdir /tmp/csiStats
for j in "${podArr[@]}"
do
	oc logs $j -n ibm-spectrum-scale-csi | grep STATS > /tmp/csiStats/$j
done

function printStats() {
		found=0
		if [ $(cat /tmp/csiStats/statlog | wc -l) -gt 1 ]; then
			tac /tmp/csiStats/statlog > /tmp/csiStats/revstatlog
			filein="/tmp/csiStats/revstatlog"
			count=0
			while IFS= read -r line
			do
				let count=count+1
				if [ "$count" == 1 ]; then
					continue
				fi
				readTime=$(echo $line | cut -d ' ' -f 2 | cut -d '.' -f 1)
				rhour=$(echo $readTime | cut -d ':' -f 1)
				rtime=$(echo $readTime | cut -d ':' -f 2)
				stime=$(echo $readTime | cut -d ':' -f 3)
				catTime=$(echo $rhour$rtime$stime)
				#echo "Concatenated time:" $catTime
				if [ "$catTime" -le "$diffTime" ]; then
					let found=1
					#echo "Found line: " $line
					oldTotalCount=$(echo $line | cut -d ' ' -f $TOTALCOUNT_POS)
					oldErrorCount=$(echo $line | cut -d ' ' -f $ERRORCOUNT_POS)
					oldMinTime=$(echo $line | cut -d ' ' -f $MINTIME_POS)
					oldMaxTime=$(echo $line | cut -d ' ' -f $MAXTIME_POS)
					oldAvgTime=$(echo $line | cut -d ' ' -f $AVGTIME_POS | cut -d '.' -f 1)
                                        floatNum=$(echo $line | cut -d ' ' -f $AVGTIME_POS | cut -d '.' -f 2)
                                        if [ "$floatNum" -ge 5 ]; then
                                                let oldAvgTime=oldAvgTime+1
                                        fi     
					oldTotalTime=$(( "$oldTotalCount" * "$oldAvgTime"))
					break
				fi
			done < "$filein"
		fi	
		lline=$(tail -1 /tmp/csiStats/statlog)	
		if test ! -z "$lline"
		then
			#echo "Latest line: " $lline
			newTotalCount=$(echo $lline | cut -d ' ' -f $TOTALCOUNT_POS)
			newErrorCount=$(echo $lline | cut -d ' ' -f $ERRORCOUNT_POS)
			newMinTime=$(echo $lline | cut -d ' ' -f $MINTIME_POS)
			newMaxTime=$(echo $lline | cut -d ' ' -f $MAXTIME_POS)
			newAvgTime=$(echo $lline | cut -d ' ' -f $AVGTIME_POS | cut -d '.' -f 1)
			floatNum=$(echo $lline | cut -d ' ' -f $AVGTIME_POS | cut -d '.' -f 2)
			if [ "$floatNum" -ge 5 ]; then
				let newAvgTime=newAvgTime+1
			fi
			if [ "$found" != 0 ]; then
				newTotalTime=$(( "$newTotalCount" * "$newAvgTime"))
				diffTotalTime=$(( "$newTotalTime" - "$oldTotalTime" ))
				diffTotalCount=$(( "$newTotalCount" - "$oldTotalCount" ))
				diffErrorCount=$(( "$newErrorCount" - "$oldErrorCount" ))
				finalAvgTime=$(( "$diffTotalTime" - "$diffTotalCount" ))
				minTime=$(echo $newMinTime "(-)")
				if [ "$newMinTime" -gt "$oldMinTime" ]; then
					minTime=$(echo $newMinTime "(" $'\U2191' ")")
				fi
				maxTime=$oldMaxTime
				if [ "$newMaxTime" -gt "$oldMaxTime" ]; then
					maxTime=$newMaxTime
				fi
				printf '    %-30s %-5s %-6s | %-13s %-13s %-13s\n' "$i"  "$diffTotalCount" "$diffErrorCount" "$finalAvgTime" "$newMinTime" "$newMaxTime"
			else
				printf '    %-30s %-5s %-6s | %-13s %-13s %-13s\n' "$i"  "$newTotalCount" "$newErrorCount" "$newAvgTime" "$newMinTime" "$newMaxTime"
			fi
		fi
	}

echo "IBM Spectrum Scale CSI Driver Detailed Statistics"
echo "Timestamp: "$(date)
echo "Showing statistics for duration "$duration" minute"
for j in "${podArr[@]}"
do
        echo ""
	node=$(oc describe pod $j -n ibm-spectrum-scale-csi | grep Node: | cut -d ' ' -f 10)
        echo "Node      : "$node
	echo "CSI POD   : "$j
	echo "Statistics:"
	if [ $(grep CSIOPStats /tmp/csiStats/$j | wc -l) -gt 0 ]; then
        echo "    CSI Operation Details                       | Operation Latency (in millseconds) " 
        echo "    ============================================|===================================="
        printf '    %-30s %-5s %-6s | %-13s %-13s %-13s\n' "Name" "Total" "Errors" "Avg" "Min" "Max"
        for i in "${csiOpStats[@]}"
        do
                grep CSIOPStats /tmp/csiStats/$j | grep $i > /tmp/csiStats/statlog
                printStats
        done
		echo ""
	fi

	echo "    REST API Details                            | API Latency (in millseconds) " 
	echo "    ============================================|===================================="
	printf '    %-30s %-5s %-6s | %-13s %-13s %-13s\n' "Name" "Total" "Errors" "Avg" "Min" "Max"
        for i in "${restAPIStats[@]}"
        do
                grep RESTAPIStats /tmp/csiStats/$j | grep $i > /tmp/csiStats/statlog
                printStats
        done
done

rm -rf /tmp/csiStats
