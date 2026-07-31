import axios from 'axios';

const TYDOM_API_BASE_URL = "/api";

export interface Device {
  deviceId: number;
  endpointId: number;
  name: string;
  type: string;
}

export class MyTydomService {

    getInfos(){
        return axios.get(TYDOM_API_BASE_URL + '/infos');
    }

    getStatus(){
        return axios.get(TYDOM_API_BASE_URL + '/status');
    }

    getDevices(){
        return axios.get(TYDOM_API_BASE_URL + '/devices');
    }

    setDeviceName(deviceId: number, endpointId: number, name: string){
        return axios.put(TYDOM_API_BASE_URL + '/devices/' + deviceId + '/' + endpointId + '/name', name)
    }

    addDevice(deviceTypeId: number){
        return axios.post(TYDOM_API_BASE_URL + '/devices/install?typeId=', deviceTypeId);
    }

    getDeviceById(deviceId: number, endpointId: number){
        return axios.get(TYDOM_API_BASE_URL + '/devices/' + deviceId + '/endpoints/' + endpointId);
    }

    deleteDevice(deviceId: number){
        return axios.delete(TYDOM_API_BASE_URL + '/device/' + deviceId);
    }
}

export default MyTydomService;

