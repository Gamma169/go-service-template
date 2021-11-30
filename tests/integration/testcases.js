// Configured in main.go
const DEFAULT_PORT = 7890;
const DB_ARR_DELIMITER = ':::';

function getPort() {    
  let portNum = process.env.FOOBAR_PORT || DEFAULT_PORT;
  return parseInt(portNum);
}

const SERVICE_URL = `http://127.0.0.1:${getPort()}`;

const USER_IDS = ["47cacb0f-94d3-490a-80a0-e9ac77fe7778", "3643e69f-f4fe-4b88-8a91-fa4a7ac9bdd8"];


const MOCK_MODELS = [
  {
    id: 'f044047d-1077-42c0-8fcd-7def81e3d9be',
    user_id: USER_IDS[0],
    name: 'some model',
    age: 21,
    some_prop: 'asdasdasd',
    some_nullable_prop: null,
    some_arr_prop:['asd','qwe','poipo', 'mnk'],
    date_created: "2021-11-20", 
    last_updated: "2021-11-27", 
  },
  {
    id: '9bc1643d-1a24-4ac2-a4c0-e59b797f7352',
    user_id: USER_IDS[0],
    name: 'another model',
    age: 34,
    some_prop: 'this is a prop',
    some_nullable_prop: 'this is not null',
    some_arr_prop:['single elem'],
    date_created: "2021-10-31", 
    last_updated: "2021-11-13", 
  },
];

function arrayToStr(arr) {
  return arr.join(DB_ARR_DELIMITER);
}

function dbSetupModels(pgClient) {
  let modelInsertString = '';
  
  MOCK_MODELS.forEach(function(model) {
    const arrStr = arrayToStr(model.some_arr_prop);
    modelInsertString = modelInsertString.concat(`
      INSERT INTO foobar_models (
        id, user_id, 
        name, age, some_prop, some_nullable_prop, some_arr_prop,
        date_created, last_updated
      ) VALUES (
        '${model.id}', '${model.user_id}', 
        '${model.name}', '${model.age}', '${model.some_prop}', '${model.some_nullable_prop}', '${arrStr}', 
        '${model.date_created}', '${model.last_updated}'
      );`);
  });

  return pgClient.query(modelInsertString);
}

const dbTeardownQuery = `DELETE FROM foobar_models;`;



module.exports = {
  SERVICE_URL,
  USER_IDS,
  MOCK_MODELS,
  arrayToStr,
  dbSetupModels,
  dbTeardownQuery,
}
