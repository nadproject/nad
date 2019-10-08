/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of Dnote.
 *
 * Dnote is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Dnote is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Dnote.  If not, see <https://www.gnu.org/licenses/>.
 */

import React, { useState, useReducer } from 'react';
import classnames from 'classnames';
import { Link } from 'react-router-dom';

import { getRepetitionsPath } from 'web/libs/paths';
import { Option, booksToOptions } from 'jslib/helpers/select';
import Modal, { Header, Body } from '../../Common/Modal';
import { useSelector } from '../../../store';
import Flash from '../../Common/Flash';
import { daysToSec } from '../../../helpers/time';
import Button from '../../Common/Button';
import MultiSelect from '../../Common/MultiSelect';
import styles from './Form.scss';
import modalStyles from '../../Common/Modal/Modal.scss';

interface Props {
  onSubmit: (formState) => void;
  cancelPath?: string;
  initialState?: FormState;
}

export interface FormState {
  title: string;
  enabled: boolean;
  hour: number;
  minute: number;
  frequency: number;
  noteCount: number;
  books: Option[];
}

enum Action {
  setTitle,
  setFrequency,
  setHour,
  setMinutes,
  setNoteCount,
  setBooks,
  toggleEnabled
}

function formReducer(state, action): FormState {
  switch (action.type) {
    case Action.setTitle:
      return {
        ...state,
        title: action.data
      };
    case Action.setFrequency:
      return {
        ...state,
        frequency: action.data
      };
    case Action.setHour:
      return {
        ...state,
        hour: action.data
      };
    case Action.setMinutes:
      return {
        ...state,
        minute: action.data
      };
    case Action.setNoteCount:
      return {
        ...state,
        noteCount: action.data
      };
    case Action.setBooks:
      return {
        ...state,
        books: action.data
      };
    case Action.toggleEnabled:
      return {
        ...state,
        enabled: !state.enabled
      };
    default:
      return state;
  }
}

const formInitialState: FormState = {
  title: '',
  enabled: true,
  hour: 8,
  minute: 0,
  frequency: daysToSec(7),
  noteCount: 20,
  books: []
};

const Form: React.FunctionComponent<Props> = ({
  onSubmit,
  cancelPath = getRepetitionsPath(),
  initialState = formInitialState
}) => {
  const [inProgress, setInProgress] = useState(false);
  const [formState, formDispatch] = useReducer(formReducer, initialState);
  const { books } = useSelector(state => {
    return {
      books: state.books.data
    };
  });
  const bookOptions = booksToOptions(books);
  const booksSelectTextId = 'book-select-text-input';

  return (
    <form
      onSubmit={e => {
        e.preventDefault();
        onSubmit(formState);
      }}
      className={styles.form}
    >
      <div className={modalStyles['input-row']}>
        <label className="input-label" htmlFor="title">
          Name
        </label>

        <input
          autoFocus
          type="text"
          id="title"
          className="text-input text-input-small text-input-stretch"
          placeholder="Weekly vocabulary reminder"
          value={formState.title}
          onChange={e => {
            const data = e.target.value;

            formDispatch({
              type: Action.setTitle,
              data
            });
          }}
        />
      </div>

      <div className={modalStyles['input-row']}>
        <label className="input-label" htmlFor={booksSelectTextId}>
          Books to include
        </label>

        <MultiSelect
          textInputId={booksSelectTextId}
          options={bookOptions}
          currentOptions={formState.books}
          setCurrentOptions={data => {
            formDispatch({ type: Action.setBooks, data });
          }}
          placeholder="Select books"
        />
      </div>

      <div
        className={classnames(
          modalStyles['input-row'],
          styles['schedule-wrapper']
        )}
      >
        <div className={styles['schedule-content']}>
          <div className={classnames(styles['schedule-input-wrapper'])}>
            <label className="input-label" htmlFor="frequency">
              How often?
            </label>

            <select
              id="frequency"
              className="form-select"
              value={formState.frequency}
              onChange={e => {
                const data = e.target.value;

                formDispatch({
                  type: Action.setFrequency,
                  data
                });
              }}
            >
              <option value={daysToSec(1)}>Every day</option>
              <option value={daysToSec(2)}>Every 2 days</option>
              <option value={daysToSec(3)}>Every 3 days</option>
              <option value={daysToSec(4)}>Every 4 days</option>
              <option value={daysToSec(5)}>Every 5 days</option>
              <option value={daysToSec(6)}>Every 6 days</option>
              <option value={daysToSec(7)}>Every week</option>
              <option value={daysToSec(14)}>Every 2 weeks</option>
              <option value={daysToSec(21)}>Every 3 weeks</option>
              <option value={daysToSec(28)}>Every 4 weeks</option>
            </select>
          </div>

          <div className={styles['schedule-input-wrapper']}>
            <label className="input-label" htmlFor="hour">
              Hour
            </label>

            <select
              id="hour"
              className={classnames('form-select', styles['time-select'])}
              value={formState.hour}
              onChange={e => {
                const data = e.target.value;

                formDispatch({
                  type: Action.setHour,
                  data
                });
              }}
            >
              {[...Array(24)].map((_, i) => {
                return (
                  <option key={i} value={i}>
                    {i}
                  </option>
                );
              })}
            </select>
          </div>

          <div className={styles['schedule-input-wrapper']}>
            <label className="input-label" htmlFor="minutes">
              Minutes
            </label>

            <select
              id="minutes"
              className={classnames('form-select', styles['time-select'])}
              value={formState.minute}
              onChange={e => {
                const data = e.target.value;

                formDispatch({
                  type: Action.setMinutes,
                  data
                });
              }}
            >
              {[...Array(60)].map((_, i) => {
                return (
                  <option key={i} value={i}>
                    {i}
                  </option>
                );
              })}
            </select>
          </div>
        </div>

        <div className={styles.help}>
          When to deliver a digest in the UTC (Coordinated Universal Time).
        </div>
      </div>

      <div className={modalStyles['input-row']}>
        <label className="input-label" htmlFor="num-notes">
          Number of notes
        </label>

        <input
          type="number"
          id="num-notes"
          className="text-input text-input-small"
          placeholder="10"
          value={formState.noteCount}
          onChange={e => {
            const data = e.target.value;

            formDispatch({
              type: Action.setNoteCount,
              data
            });
          }}
        />

        <div className={styles.help}>
          Maximum number of notes to include in each repetition
        </div>
      </div>

      <div className={modalStyles['input-row']}>
        <label className="input-label" htmlFor="enabled">
          Enabled?
        </label>

        <div>
          <input
            type="checkbox"
            id="enabled"
            checked={formState.enabled}
            onChange={e => {
              const data = e.target.value;

              formDispatch({
                type: Action.toggleEnabled,
                data
              });
            }}
          />
        </div>
      </div>

      <div className={modalStyles.actions}>
        <Button type="submit" kind="first" size="normal" isBusy={inProgress}>
          Create
        </Button>

        <Link
          to={cancelPath}
          onClick={e => {
            const ok = window.confirm('Are you sure?');
            if (!ok) {
              e.preventDefault();
              return;
            }
          }}
          className="button button-second button-normal"
        >
          Cancel
        </Link>
      </div>
    </form>
  );
};

export default Form;
