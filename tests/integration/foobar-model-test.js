const chai = require('chai');
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

const { Client } = require('pg');

// Make sure you require the health check endpoint test so that it runs first
require('./health-check-test');

const {
  SERVICE_URL,
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
    testsPGClient.connect().catch(function() {
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
          .get(`/user/foobar-models`)
          .set('user-id', id)
          .then(function(resp) {
            
            console.log(resp.body);
            done();


          }, done)
          .catch(done);
        });

        it.skip('should be able to get jsonapi models for the user:' + id, function(done) {
          // TODO
          done();
        });

      });

    });


    describe('Post Model Tests', function() {
    });


    describe('Update Model Tests', function() {
    });


    describe('Delete Model Tests', function() {
    });


    // End of Endpoints Tests describe
  });
  // End of Test Suite describe
});
