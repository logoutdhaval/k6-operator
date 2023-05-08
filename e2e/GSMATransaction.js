import http from 'k6/http';
import { sleep } from 'k6';
import {group ,check } from 'k6';
const transactionIds = []
    //var transactionComplete = 0;
    //var transactionFailed = 0;
    //var transactionInProgress = 0;
export const options = {
  insecureSkipTLSVerify: true,
  vus: 5,
  duration: '1h',
  iterations: 10,
};
export default function () {
    gsmaTransaction();
    sleep(Math.floor(Math.random() * 3) + 1)
    if(transactionIds.length < `${options.iterations}`){
        getStatus();
    }
}
export function gsmaTransaction(){
  const url = 'https://channel.sandbox.fynarfin.io/channel/gsma/transaction';
  const payload = JSON.stringify({
                                     "requestingOrganisationTransactionReference": "string",
                                     "subType": "inbound",
                                     "type": "transfer",
                                     "amount": "11",
                                     "currency": "USD",
                                     "descriptionText": "string",
                                     "requestDate": "2022-09-28T12:51:19.260+00:00",
                                     "customData": [
                                         {
                                             "key": "string",
                                             "value": "string"
                                         }
                                     ],
                                     "payer": [
                                         {
                                             "partyIdType": "MSISDN",
                                             "partyIdIdentifier": "+44999911"
                                         }
                                     ],
                                     "payee": [
                                         {
                                             "partyIdType": "accountId",
                                             "partyIdIdentifier": "L000000001"
                                         }
                                     ]
                                 });

  const params = {
    headers: {
        "X-CorrelationID":"123456789",
        "accountHoldingInstitutionId":"gorilla",
        "amsName":"mifos",
        "Content-Type":"application/json",
    },
  };
  var res = http.post(url, payload, params);
  transactionIds.push(res.json().transactionId)
};

export function getStatus(){
  const params = {
    headers: {
      "Platform-TenantId":"gorilla"
    },
  };
  for(var index = transactionIds.length-1; index >=0 ; index--){
    var url = 'https://ops-bk.sandbox.fynarfin.io/api/v1/transfers?size=1&page='+index;
//    sleep(3)
    var res = http.get(url, params);
//    console.log("Gsma transaction id: res.json().content[].id = ",res.json().content[0].id);
    check("COMPLETED", {
                'Check for success rate of GSMA Transaction': (r) => r == res.json().content[0].status
              });

  }
}