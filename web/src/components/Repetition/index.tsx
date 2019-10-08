import React, { useState, useEffect } from 'react';
import Helmet from 'react-helmet';
import { Link } from 'react-router-dom';
import classnames from 'classnames';

import { getNewRepetitionPath } from 'web/libs/paths';
import { getDigestRules } from '../../store/repetitionRules';
import { useDispatch } from '../../store';
import styles from './Repetition.scss';

const Repetition: React.FunctionComponent = () => {
  const dispatch = useDispatch();
  useEffect(() => {
    dispatch(getDigestRules());
  }, [dispatch]);

  return (
    <div className="page page-mobile-full">
      <Helmet>
        <title>Repetition</title>
      </Helmet>

      <div className="container mobile-fw">
        <div className={classnames('page-header', styles.header)}>
          <h1 className="page-heading">Repetition</h1>

          <Link
            className="button button-first button-normal"
            to={getNewRepetitionPath()}
          >
            New
          </Link>
        </div>
      </div>

      <div className="container">content</div>
    </div>
  );
};

export default Repetition;
