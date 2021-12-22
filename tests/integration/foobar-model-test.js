const chai = require('chai');
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

const { Client } = require('pg');

// Make sure you require the health check endpoint test so that it runs first
require('./health-check-test');

const {
  SERVICE_URL,
  REQUESTER_ID_HEADER,
  USER_IDS,
  MOCK_MODELS,
  // arrayToStr,
  dbSetupModels,
  dbTeardownQuery,
} = require('./testcases.js');


// From ./scripts/setup_database
const DATABASE_NAME = 'foo';

let testsPGClient;

describe('foobar Model Tests:', function() {

  before('Setup Database Connections', function(done) {
    const port = parseInt(process.env.DATABASE_PORT || '5432')
    testsPGClient = new Client({
      user: 'postgres',
      host: 'localhost',
      database: DATABASE_NAME,
      port,
    });
    testsPGClient.connect().catch(function(err) {
      console.error(err)  // eslint-disable-line
      console.error("ERROR: Connection to postgres not established -- Check your docker container and port mappings.  Make sure it's running on port: " + port); // eslint-disable-line
      testsPGClient.end();
      process.exit(1);
    });
    setTimeout(done, 250);
  });

  after('Shutdown Database Connections', function(done) {
    testsPGClient.end();
    setTimeout(done, 250);
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

                chai.assert.deepInclude(returnedModel, {
                  id: mockModel.id,
                  name: mockModel.name,
                  age: mockModel.age,
                  someProp: mockModel.some_prop,
                  someNullableProp: mockModel.some_nullable_prop,
                  someArrProp: mockModel.some_arr_prop,
                }, "models match");

                // golang json encoding time field returns much more precision than what I provide in the mocks so we do date checking seperately
                chai.assert.include(returnedModel.dateCreated, mockModel.date_created, "mock dateCreated is in returned date");
                chai.assert.include(returnedModel.lastUpdated, mockModel.last_updated, "mock lastUpdated is in returned date");
                
              });

              // Test is sucessful
              done();

            }, done)
            .catch(done);
        });


        it('should be able to get jsonapi models for the user:' + id, function(done) {
          chai.request(SERVICE_URL)
          .get('/user/foobar-models')
          .set(REQUESTER_ID_HEADER, id)
          .set('Accept', 'application/vnd.api+json')
          .then(function(resp) {
            
            const returnedModels = resp.body;
            chai.assert.equal(returnedModels.data.length, modelsForId.length, "returns same number of models as in testcases");
            modelsForId.forEach(mockModel => {
              const returnedModel = returnedModels.data.find(retModel => retModel.id === mockModel.id);

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
              }, "models match");

            });

            // Test is sucessful
            done();

          }, done)
          .catch(done);
        });

      });

    });


    describe('Post Model Tests', function() {
    });


    describe('Update Model Tests', function() {
    });


    describe('Delete Model Tests', function() {
      const modelToDelete = MOCK_MODELS[0];

      it('should be able to delete a model that belongs to the requester', function(done) {
        chai.request(SERVICE_URL)
          .delete('/user/foobar-models/'+ modelToDelete.id)
          .set(REQUESTER_ID_HEADER, modelToDelete.user_id)
          .then(function() {
            testsPGClient.query(`SELECT id FROM foobar_models`)
              .then(function(resp){
                chai.assert.ok(resp.rows);
                chai.assert.equal(MOCK_MODELS.length-1, resp.rows.length);
                const nonDeletedModelIds = MOCK_MODELS.filter(model => model.id !== modelToDelete.id).map(m => m.id);
                const returnedIds = resp.rows.map(m => m.id);
                chai.assert.sameMembers(nonDeletedModelIds, returnedIds);
                done();
              })
              .catch(done);
          }, done)
          .catch(done);
      });

      it('should NOT delete an analytics file that does not belong to the owner', function(done) {
        const modelOwnerId = modelToDelete.user_id;
        const badClientId = USER_IDS.filter(id => id !== modelOwnerId)[0];
        chai.request(SERVICE_URL)
          .delete('/user/foobar-models/'+ modelToDelete.id)
          .set(REQUESTER_ID_HEADER, badClientId)
          .then(function() {
            testsPGClient.query(`SELECT id FROM foobar_models`)
              .then(function(resp){
                chai.assert.ok(resp.rows);
                chai.assert.equal(MOCK_MODELS.length, resp.rows.length);
                done();
              })
              .catch(done);
          }, done)
          .catch(done);
      });
    });


    // End of Endpoints Tests describe
  });
  // End of Test Suite describe
});
