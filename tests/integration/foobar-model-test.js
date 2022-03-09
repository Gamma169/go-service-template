const chai = require('chai');
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

const { Client } = require('pg');

const { v4: uuidv4 } = require('uuid');

// Make sure you require the health check endpoint test so that it runs first
require('./health-check-test');

const {
  SERVICE_URL,
  DATABASE_NAME,
  REQUESTER_ID_HEADER,
  USER_IDS,
  MOCK_MODELS,
  MOCK_SUB_MODELS,
  arrayToStr,
  dbSetupModels,
  dbTeardownQuery,
} = require('./testcases.js');

let testsPGClient;

describe('foobar Model Tests:', function() {

  before('Setup Database Connections', function(done) {
    const port = parseInt(process.env.DATABASE_PORT || '5432');
    testsPGClient = new Client({
      user: 'postgres',
      host: 'localhost',
      database: DATABASE_NAME,
      port,
    });
    testsPGClient.connect()
      .then(() => done())
      .catch(function() {
        testsPGClient.end();
        done(new Error("Connection to postgres not established -- Check your docker container and port mappings.  Make sure it's running on port: " + port));
      });
  });

  after('Shutdown Database Connections', function(done) {
    testsPGClient
      .end()
      .finally(() => done());
  });


  describe('foobar Model Endpoint Tests', function() {

    beforeEach('Populate Database', function(done) {
      dbSetupModels(testsPGClient)
        .then(function() {
          done();
        })
        .catch(done);
    });

    afterEach('Teardown Database', function(done) {
      testsPGClient
        .query(dbTeardownQuery)
        .then(function() {
          done();
        })
        .catch(done);
    });

    describe('Get Model Tests', function() {

      USER_IDS.forEach(id => {
        const modelsForId = MOCK_MODELS.filter(model => model.user_id === id);

        it('should be able to get json models for the user: ' + id, function(done) {
          chai.request(SERVICE_URL)
            .get('/user/foobar-models')
            .set(REQUESTER_ID_HEADER, id)
            .then(function(resp) {

              const returnedModels = resp.body;
              chai.assert.equal(returnedModels.length, modelsForId.length, "returns same number of models as in testcases");
              modelsForId.forEach(mockModel => {
                const returnedModel = returnedModels.find(retModel => retModel.id === mockModel.id);
                const subModels = MOCK_SUB_MODELS.filter(model => model.foobar_model_id === returnedModel.id)
                  .map(s => ({id: s.id, foobarModelId: s.foobar_model_id, value: s.value, valueInt: s.value_int, __id__: ""}));

                chai.assert.deepInclude(returnedModel, {
                  id: mockModel.id,
                  name: mockModel.name,
                  age: mockModel.age,
                  someProp: mockModel.some_prop,
                  someNullableProp: mockModel.some_nullable_prop,
                  someArrProp: mockModel.some_arr_prop,
                  subModels: subModels,
                }, "models match");

                // golang json encoding time field returns much more precision than what I provide in the mocks so we do date checking seperately
                chai.assert.include(returnedModel.dateCreated, mockModel.date_created, "mock dateCreated is in returned date");
                chai.assert.include(returnedModel.lastUpdated, mockModel.last_updated, "mock lastUpdated is in returned date");

              });

              // Test is sucessful
              done();

            })
            .catch(done);
        });


        it('should be able to get jsonapi models for the user:' + id, function(done) {
          chai.request(SERVICE_URL)
            .get('/user/foobar-models')
            .set(REQUESTER_ID_HEADER, id)
            .set('Accept', 'application/vnd.api+json')
            .then(function(resp) {

              const returnedModels = resp.body;
              const included = returnedModels.included;
              chai.assert.equal(returnedModels.data.length, modelsForId.length, "returns same number of models as in testcases");
              modelsForId.forEach(mockModel => {
                const returnedModel = returnedModels.data.find(retModel => retModel.id === mockModel.id);
                const subModels = MOCK_SUB_MODELS.filter(model => model.foobar_model_id === returnedModel.id);
                const relationships = {
                  'sub-models': {
                    data: subModels.map(s => (
                      {id: s.id, type: 'sub-models'}
                    )),
                  },
                };

                chai.assert.deepEqual(returnedModel, {
                  attributes: {
                    'name': mockModel.name,
                    'age': mockModel.age,
                    'some-prop': mockModel.some_prop,
                    'some-nullable-prop': mockModel.some_nullable_prop,
                    'some-arr-prop': mockModel.some_arr_prop,
                    // jsonapi library in golang returns seconds since 1970.  getTime returns mS, so slight convertion is necessary
                    'date-created': new Date(mockModel.date_created).getTime()/1000,
                    'last-updated': new Date(mockModel.last_updated).getTime()/1000,
                    // This is only necessary for ember-save-relationships mixin
                    '__id__': "",
                  },
                  id: mockModel.id,
                  type: 'foobar-model',
                  relationships,
                }, "models match");

                if (subModels.length > 0) {
                  chai.assert.equal(included.length, subModels.length);

                  const expectedIncluded = subModels.map(s => (
                    {
                      id: s.id,
                      type: "sub-models",
                      attributes: {
                        'foobar-model-id': s.foobar_model_id,
                        'value': s.value,
                        'value-int': s.value_int,
                        '__id__': '',
                      }
                    }
                  ));

                  // We use `sameDeepMembers` because included array can be in any order
                  chai.assert.sameDeepMembers(included, expectedIncluded, "included matches");
                }
              });

              // Test is sucessful
              done();
            })
            .catch(done);
        });

      });

    });


    describe('Post Model Tests', function() {

      /********* TEST HELPERS *********/

      const generalInput = {
        name: 'some post model',
        age: 44,
        someProp: 'qweqweqwe',
        someNullableProp: null,
        someArrProp:['zxc','sdf','xcv', 'popop'],
        __id__: uuidv4(),
      };

      function testFoobarModelDatabaseInput(pgResp, newModelId) {
        chai.assert.equal(1, pgResp.rows.length);

        const expectedRow = {
          id: newModelId,
          user_id: USER_IDS[1],
          name: generalInput.name,
          age: generalInput.age,
          some_prop: generalInput.someProp,
          some_nullable_prop: generalInput.someNullableProp,
          some_arr_prop: arrayToStr(generalInput.someArrProp),
        };

        chai.assert.deepEqual(pgResp.rows[0], expectedRow);
      }

      /********* TESTS *********/

      describe('JSON Input Format', function() {

        /********* TEST HELPERS *********/

        function makeRequest(input) {
          return chai.request(SERVICE_URL)
            .post('/user/foobar-models')
            .set('user-id', USER_IDS[1])
            .set('Content-Type', 'application/json')
            .send(input);
        }

        function testServerResponse(serverResp) {
          const newModel = serverResp.body;
          const newModelId = newModel.id;

          chai.assert.deepInclude(newModel, generalInput);
          return newModelId;
        }

        /********* TESTS *********/

        describe('WITHOUT Sub Models', function() {
          it('should be able to add new model WITHOUT id + WITHOUT Sub Models to the database using json', function(done) {
            let newModelId;
            makeRequest(generalInput)
              .then(function(serverResp) {
                newModelId = testServerResponse(serverResp);
                chai.assert.deepEqual(serverResp.body.subModels, []);
                return testsPGClient.query(`SELECT 
                    id, user_id, name, age, some_prop, some_nullable_prop, some_arr_prop
                  FROM foobar_models WHERE id = '${newModelId}'`);
              })
              .then(function(pgResp) {
                testFoobarModelDatabaseInput(pgResp, newModelId);
                done();
              })
              .catch(done);
          });

          it('should be able to add new model WITH id + WITHOUT Sub Models to the database using json', function(done) {
            const modelId = uuidv4();
            const input = Object.assign({id: modelId}, generalInput);
            makeRequest(input)
              .then(function(serverResp) {
                testServerResponse(serverResp, []);
                chai.assert.deepEqual(serverResp.body.subModels, []);
                return testsPGClient.query(`SELECT 
                    id, user_id, name, age, some_prop, some_nullable_prop, some_arr_prop
                  FROM foobar_models WHERE id = '${modelId}'`);
              })
              .then(function(pgResp) {
                testFoobarModelDatabaseInput(pgResp, modelId);
                done();
              })
              .catch(done);
          });

        });

        describe('INCLUDING Sub Models', function() {

          /********* TEST HELPERS *********/

          const subModels = [
            {
              foobarModelId: '',
              value: uuidv4(),
              valueInt: Math.floor(Math.random() * 100),
              __id__: uuidv4(),
            },
            {
              foobarModelId: '',
              value: uuidv4(),
              valueInt: Math.floor(Math.random() * 100),
              __id__: uuidv4(),
            }
          ];
          const input = Object.assign({subModels}, generalInput);

          function testSubModelResponse(serverResp, newModelId) {
            const subModels = serverResp.body.subModels;
            const newIds = subModels.map(s => s.id);
            const expectedSubModelResp = subModels.map((s, i) => Object.assign({}, s, {id: newIds[i], foobarModelId: newModelId}));
            chai.assert.sameDeepMembers(subModels, expectedSubModelResp);
            return newIds;
          }

          function testSubModelDatabaseResponse(pgResp, newSubModelIds) {
            chai.assert.equal(pgResp.rows.length, subModels.length);

            const expectedPgResp = subModels.map((s, idx) => ({
              id: newSubModelIds[idx],
              user_id: USER_IDS[1],
              value: subModels[idx].value,
              value_int: subModels[idx].valueInt,
            }));

            chai.assert.deepEqual(pgResp.rows, expectedPgResp);
          }

          /********* TESTS *********/

          it('should be able to add new model WITHOUT id + INCLUDING Sub Models to the database using json', function(done) {
            let newModelId;
            let newSubModelIds;
            makeRequest(input)
              .then(function(serverResp) {
                newModelId = testServerResponse(serverResp);
                newSubModelIds = testSubModelResponse(serverResp, newModelId);
                return testsPGClient.query(`SELECT 
                    id, user_id, name, age, some_prop, some_nullable_prop, some_arr_prop
                  FROM foobar_models WHERE id = '${newModelId}'`);
              })
              .then(function(pgResp) {
                testFoobarModelDatabaseInput(pgResp, newModelId);
                return testsPGClient.query(`SELECT 
                    id, user_id, value, value_int
                  FROM sub_models WHERE foobar_model_id = '${newModelId}'`);
              })
              .then(function(pgResp) {
                testSubModelDatabaseResponse(pgResp, newSubModelIds);
                done();
              })
              .catch(done);
          });

          it('should be able to add new model WITH id + INCLUDING Sub Models to the database using json', function(done) {
            const modelId = uuidv4();
            const subModelsWithIds = subModels.map(s => Object.assign({}, s, {id: uuidv4(), foobarModelId: modelId}));
            const subModelIds = subModelsWithIds.map(s => s.id);
            const input = Object.assign({id: modelId, subModels: subModelsWithIds}, generalInput);
            makeRequest(input)
              .then(function(serverResp) {
                testServerResponse(serverResp);
                testSubModelResponse(serverResp, modelId);
                return testsPGClient.query(`SELECT 
                    id, user_id, name, age, some_prop, some_nullable_prop, some_arr_prop
                  FROM foobar_models WHERE id = '${modelId}'`);
              })
              .then(function(pgResp) {
                testFoobarModelDatabaseInput(pgResp, modelId);
                return testsPGClient.query(`SELECT 
                    id, user_id, value, value_int
                  FROM sub_models WHERE foobar_model_id = '${modelId}'`);
              })
              .then(function(pgResp) {
                testSubModelDatabaseResponse(pgResp, subModelIds);
                done();
              })
              .catch(done);
          });

        });

      });

      describe('JSONAPI Input Format', function() {

        describe('WITHOUT Sub Models', function() {
          // TODO
          it.skip('should be able to add new model WITHOUT id + WITHOUT Sub Models to the database using jsonAPI', function(done) {
            done();
          });

          it.skip('should be able to add new model WITH id + WITHOUT Sub Models to the database using jsonAPI', function(done) {
            done();
          });
        });

        describe('INCLUDING Sub Models', function() {
          it.skip('should be able to add new model WITHOUT id + INCLUDING Sub Models to the database using jsonAPI', function(done) {
            done();
          });

          it.skip('should be able to add new model WITH id + INCLUDING Sub Models to the database using jsonAPI', function(done) {
            done();
          });
        });

      });

    });


    describe('Update Model Tests', function() {
      // TODO
    });


    describe('Delete Model Tests', function() {
      const modelToDelete = MOCK_MODELS[0];

      it('should be able to delete a model INCLUDING Sub Models that belongs to the requester', function(done) {
        // Make sure model deleting has sub models in order to make sure we delete them as well
        const subModelsToDelete = MOCK_SUB_MODELS.filter(s => s.foobar_model_id === modelToDelete.id);
        chai.assert.isAbove(subModelsToDelete.length, 0, "model to delete should have sub models associated with it");

        chai.request(SERVICE_URL)
          .delete('/user/foobar-models/'+ modelToDelete.id)
          .set(REQUESTER_ID_HEADER, modelToDelete.user_id)
          .then(function() {
            return testsPGClient.query(`SELECT id FROM foobar_models`);
          })
          .then(function(pgResp) {
            chai.assert.ok(pgResp.rows);
            chai.assert.equal(MOCK_MODELS.length-1, pgResp.rows.length);
            const nonDeletedModelIds = MOCK_MODELS.filter(model => model.id !== modelToDelete.id).map(m => m.id);
            const returnedIds = pgResp.rows.map(m => m.id);
            chai.assert.sameMembers(nonDeletedModelIds, returnedIds);
            return testsPGClient.query(`SELECT id FROM sub_models WHERE foobar_model_id = '${modelToDelete.id}'`);
          })
          .then(function(pgResp) {
            chai.assert.ok(pgResp.rows);
            chai.assert.equal(pgResp.rows.length, 0);
            done();
          })
          .catch(done);
      });

      it('should NOT delete a model that does not belong to the owner', function(done) {
        const modelOwnerId = modelToDelete.user_id;
        const badClientId = USER_IDS.filter(id => id !== modelOwnerId)[0];
        chai.request(SERVICE_URL)
          .delete('/user/foobar-models/'+ modelToDelete.id)
          .set(REQUESTER_ID_HEADER, badClientId)
          .then(function() {
            return testsPGClient.query(`SELECT id FROM foobar_models`);
          })
          .then(function(pgResp) {
            chai.assert.ok(pgResp.rows);
            chai.assert.equal(MOCK_MODELS.length, pgResp.rows.length);
            done();
          })
          .catch(done);
      });
    });


    // End of Endpoints Tests describe
  });
  // End of Test Suite describe
});
