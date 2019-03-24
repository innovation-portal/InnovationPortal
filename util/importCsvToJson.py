# -*- coding: utf-8 -*-
"""
Created on Sat Mar 23 14:03:48 2019

    Extract CSV data, transform key/value to match desired server keys and value types, load by posting JSON
Parse CSV data and create JSON
    Print JSON if desired
    POST JSON to endpoint to load backend DB

@author: Owner
"""
import csv
import requests
import json 
import sys
import re

default_import_filename='HackPortal.csv'
default_url="http://104.197.161.63:8080/v1/api/projects"
import_filename=default_import_filename

# airfields left here to help with maintaining mapping to clumn names in airtable spreadsheet fields
# airfields=['Name', 'HackathonProgram', 'Sector', 'MonthYear', 'WinnerType', 'Type', 'Names', 'Descriptions']
mongofields=['Name', 'Hackathon', 'ApplicationArea', 'Year',  'Tags', 'Members', 'Description','WinnerType','Twitter']
arrayFields=['ApplicationArea', 'Tags', 'Members']

fields=mongofields
debug=False # True

def getJsonArray():
    payload_items=[]
    try:
        with open(import_filename, encoding='utf-8') as csvfile:
            readCsv =csv.reader(csvfile,delimiter=',')
            cnt=0
            for row in readCsv:
                idx=0
                payload={}
                for utfitem in row:
                    item=re.sub(r'[^\x00-\x7f]',r'', utfitem)
                    # handle case if new fields added to CSV
                    if idx >= len(fields):
                        continue
                    if 'Year' in fields[idx]:
                        # Year is a special case - parse and ensure its an int
                        intitem=handleYearCase(item)
                        payload[fields[idx]] = intitem
                    elif fields[idx] in arrayFields:
                        item_array=[]
                        if item:
                            item_array=item.split(sep=",")
                        payload[fields[idx]] = item_array
                    else:
                        payload[fields[idx]] = item
                    idx+=1
                cnt+=1
                payload_items.append(payload)
    except FileNotFoundError:
        print ("Exiting with Error: File not found.  Expected file: %s"%(import_filename))
        sys.exit()
    if debug:
       print ("Rows: %d"%(cnt-1))
    return payload_items

def handleYearCase(item):
    if len(item) > 4:
        item=item[-4:]
        if item.isdecimal():
            intitem = int(item)
        else:
            intitem = 0
    else:
        intitem = 0
    return intitem

def showJsonArray():
    cnt=0
    payload_items=[]
    with open(import_filename, encoding='utf-8') as csvfile:
        readCsv =csv.reader(csvfile,delimiter=',')
        print("[")
        firstRow=True
        for row in readCsv:
            idx=0
            # skip zero row which has field names
            cnt+=1
            
            if not firstRow:
                print(",")
            else:
                firstRow=False
            payload={}
            print("{")
            for utfitem in row:
                item=re.sub(r'[^\x00-\x7f]',r'', utfitem)
                if idx >= len(fields):
                    continue
                if 'Year' in fields[idx]:
                    #print('Year')
                    if len(item) > 4:
                        item=item[-4:]
                        if item.isdecimal():
                            intitem = int(item)
                        else:
                            intitem = 0
                    else:
                        intitem = 0
                    print ("'{}':{},".format(fields[idx],item))
                    payload[fields[idx]] = intitem
                elif fields[idx] in arrayFields:
                    item_array=[]
                    if item:
                        item_array=item.split(sep=",")
                    print ("'{}':{},".format(fields[idx],item_array))
                    payload[fields[idx]] = item_array
                else:
                    print ("'{}':'{}',".format(fields[idx],item))                 
                    payload[fields[idx]] = item
                idx+=1
            print("}", end="")
            payload_items.append(payload)
            
            #if (cnt>4):
            #    break
        print("]")
        print ("Rows: %d"%(cnt-2))
        return payload_items
    

       
def sendJsonPost(payload):
    myheaders = {'content-type': 'application/json'}
    result = requests.post(url, data=json.dumps(payload), headers=myheaders)
    print ("Result: {} from url {} with  posting: {}".format(result,url,payload))
        
def getServerResponse():
    result=requests.get(url)
    data = result.json()
    print ("Get Result: {} from url {} with get returning {}".format(result,url, json.dumps(data)))

def showPayloadItemArray():
    result=showJsonArray()
    print ("======= RESULTS =======")
    for item in result:
        print (item)
    
        
def sendJsonPost(payload, url):
    print("sending to url  %s"%(url))
    myheaders = {'content-type': 'application/json'}
    result = requests.post(url, data=json.dumps(payload), headers=myheaders)
    print ("Result: {} from url {} with  posting: {}".format(result,url,payload))
    
def postItems(url):
    result=getJsonArray()
    print ("======= RESULTS =======")
    idx=0
    for item in result:
        sendJsonPost(item, url)
        idx+=1
    
if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Usage:  importCsvToJson <FileToLoad> <UrlToPostJson>")  
        print("Usage:  use default to replace either or both parameters")
        print("Usage:  importCsvToJson default default -- to use defaults for both")
        sys.exit()
    import_filename = sys.argv[1]
    if ("default" == import_filename):
        import_filename=default_import_filename
        print ("Using default import file [%s]"%(default_import_filename))

    url = sys.argv[2]
    if ("default" == url):
        url=default_url
        print ("Using default url [%s]"%(default_url))
    postItems(url)

    
